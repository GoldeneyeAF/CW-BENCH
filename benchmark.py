import openai

from tqdm.auto import tqdm

import os
import json
import re
import unittest
import time
import sys
from typing import Dict, List, Tuple
import subprocess
import uuid
import io
from contextlib import redirect_stdout

# Конфигурация
TEST_LIBRARIES = {
    "python": "unittest",
    "go": "testing"
}

api = openai.OpenAI(
    base_url="http://host.docker.internal:11434/v1",
    api_key="fake",
)

SEPARATOR = "### INCORRECT TESTS ###"
MODEL_NAMES = ["llama3.2:3b-instruct-q8_0"]

class Benchmark:
    def __init__(self, tasks_root: str, model_name):
        self.tasks_root = tasks_root
        self.results = {
            "unit": {"executable": 0, "mutated": 0, "total": 0, "spamy": 0},
            "functional": {"passed": 0, "total": 0},
            "integrational": {"correct": 0, "incorrect": 0, "total_correct": 0, "total_incorrect": 0}
        }
        self.model_name = model_name

    def run(self):
        # Проходим по всем типам заданий
        for test_type in ["unit", "integrational", "functional"]:
            for language in ["python", "go"]:
                test_dir = os.path.join(self.tasks_root, f"{test_type}")
                if not os.path.exists(test_dir):
                    continue

                # Обрабатываем каждое задание
                for task_name in os.listdir(test_dir):
                    if not task_name.startswith(language):
                        continue

                    task_path = os.path.join(test_dir, task_name)
                    print(task_path)
                    if not os.path.isdir(task_path):
                        continue

                    # Обрабатываем все файлы заданий
                    for task_file in os.listdir(task_path):
                        if not task_file.endswith("code_samples.json"):
                            continue

                        full_path = os.path.join(task_path, task_file)
                        task_data = self._load_task_data(full_path, test_type, language)

                        if not task_data:
                            continue

                        for task in task_data:
                            # Получаем тесты от LLM
                            filename = f"temp_{uuid.uuid4().hex}"
                            test_code = self._get_tests_from_llm(task, test_type, language, filename)

                            if test_code == "":
                                continue

                            test_code = test_code.replace("```python", "")
                            test_code = test_code.replace("```go\n", "")

                            test_code = test_code.replace("```", "")

                            # Запускаем тесты и собираем результаты
                            #test_code = ""
                            self._run_tests_and_collect_results(test_code, task, test_type, language, filename)

        # Выводим финальные результаты
        self._print_final_results()

    def _load_task_data(self, file_path: str, test_type: str, language: str) -> Dict:
        with open(file_path, 'r') as f:
            full_data = json.load(f)

        final_data = []

        for data in full_data:
            task_data = {"id": data["id"], "type": test_type, "language": language}

            if test_type == "unit":
                if language == "python":
                    task_data.update({
                        "code": data["code"],
                        "mutated_code": data["mutated_code"],
                        "function": data["function_name"]
                    })
                else:  # go
                    with open(os.path.join(os.path.dirname(file_path), data["code_path"] + ".go"), 'r') as f:
                        code = f.read()
                    with open(os.path.join(os.path.dirname(file_path), data["mutated_path"] + ".go"), 'r') as f:
                        mutated_code = f.read()

                    task_data.update({
                        "code": code,
                        "mutated_code": mutated_code,
                        "function": data["function_name"]
                    })

            elif test_type == "integrational":
                if language == "python":
                    task_data["code"] = data["code"]
                else:  # go
                    with open(os.path.join(os.path.dirname(file_path), data["code_path"] + ".go"), 'r') as f:
                        task_data["code"] = f.read()

            else:  # functional
                task_data.update({
                    "code": data["full_code"],
                    "task": data["desription"]
                })

            final_data.append(task_data)

        return final_data

    def _get_tests_from_llm(self, task_data: Dict, test_type: str, language: str, filename) -> str:
        if test_type == "unit":
            prompt = self._create_unit_prompt(task_data["code"], task_data["function"], language, filename)
        elif test_type == "integrational":
            prompt = self._create_integrational_prompt(task_data["code"], language, filename)
        else:  # functional
            prompt = self._create_functional_prompt(task_data["task"], language, filename)

        # Реальный вызов LLM API
        try:
            response = api.chat.completions.create(
                model=self.model_name,
                messages=[{
                    "role": "user",
                    "content": prompt
                }],
                temperature=0.3
            )
            return response.choices[0].message.content


        except Exception as e:
            print(e)
            return ""

    def _create_unit_prompt(self, code: str, function: str, language: str, filename) -> str:
        test_lib = TEST_LIBRARIES[language]
        return f"""
Write comprehensive unit tests for the following {language} code. The tests should be written using {test_lib} library.
Focus specifically on testing the function '{function}'. You can call import function for test from {filename}
The tests should: Output detailed error messages when assertions fail. Do not write more than 30 tests
Provide only the test code without any additional explanations.

Code to test:
{code}
"""

    def _create_integrational_prompt(self, code: str, language: str, filename) -> str:
        test_lib = TEST_LIBRARIES[language]
        return f"""
Write unit tests for the following {language} code. The tests should be written using {test_lib} library.
Provide two sets of tests separated by '{SEPARATOR}':
1. Tests that should pass (correct scenarios)
2. Tests that should fail (incorrect scenarios)
You should import function for test only from {filename}.
Provide only the test code without any additional explanations.

Code to test:
{code}
"""

    def _create_functional_prompt(self, task: str, language: str, filename) -> str:
        test_lib = TEST_LIBRARIES[language]
        return f"""
Write functional tests based on the following requirements. The tests should be written using {test_lib} library.
You can call import function for test from {filename}.
The tests should:
1. Verify all functional requirements
2. Include edge cases
3. Output detailed error messages when assertions fail
Provide only the test code without any additional explanations.

Requirements:
{task}
"""

    def _run_tests_and_collect_results(self, test_code: str, task_data: Dict, test_type: str, language: str, filename):
        if test_type == "unit":
            self._run_unit_tests(test_code, task_data, language, filename)
        elif test_type == "integrational":
            self._run_integrational_tests(test_code, task_data, language, filename)
        else:  # functional
            self._run_functional_tests(test_code, task_data, language, filename)

    def _run_unit_tests(self, test_code: str, task_data: Dict, language: str, filename):
        # Сначала тестируем исходный код
        task_code = task_data['code']
        original_result = self._execute_code(task_code, test_code, language, filename)
        time.sleep(0.5)
        # Затем тестируем мутированный код
        task_code = task_data['mutated_code']
        mutated_result = self._execute_code(task_code, test_code, language, filename, mut=True)

        # Сравниваем выводы и обновляем результаты
        if 'all' in original_result:
            self.results["unit"]["total"] += original_result['all']

            # Executable: тесты должны скомпилироваться и запуститься
            self.results["unit"]["executable"] += original_result['all'] - original_result['errors']

            self.results["unit"]["spamy"] += original_result['failed']
        if 'all' in mutated_result:
            # Mutated: тесты должны обнаружить различия между оригинальным и мутированным кодом
            if original_result['errors'] != mutated_result['errors'] or original_result['failed'] != mutated_result['failed']:
                self.results["unit"]["mutated"] += original_result['all'] - original_result['errors']

    def _run_integrational_tests(self, test_code: str, task_data: Dict, language: str, filename):
        # Разделяем тесты на правильные и неправильные
        try:
            correct_tests, incorrect_tests = test_code.split(SEPARATOR)
        except ValueError:
            return
        # Тестируем правильные тесты
        task_code = task_data['code']
        correct_result = self._execute_code(task_code, correct_tests, language, filename)
        # Тестируем неправильные тесты
        incorrect_result = self._execute_code(task_code, incorrect_tests, language, filename)
        # Обновляем результаты
        if 'all' in correct_result:
            self.results["integrational"]["total_correct"] += correct_result['all']
            self.results["integrational"]["correct"] += correct_result['all'] - correct_result['errors'] - correct_result['failed']
        if 'all' in incorrect_result:
            self.results["integrational"]["total_incorrect"] += incorrect_result['all']
            if correct_result['all'] - correct_result['errors'] - correct_result['failed'] != 0:
                self.results["integrational"]["incorrect"] += incorrect_result['failed'] + incorrect_result['errors']

    def _run_functional_tests(self, test_code: str, task_data: Dict, language: str, filename):
        # Тестируем функциональные тесты
        task_code = task_data['code']
        result = self._execute_code(task_code, test_code, language, filename)
        # Обновляем результаты
        if 'all' in result:
            self.results["functional"]["total"] += result['all']
            self.results["functional"]["passed"] += result['all'] - result['errors'] - result['failed']

    def _execute_code(self, task_code: str, test_code: str, language: str, filename: str, mut=False) -> Dict:
        # Создаем временный файл
        temp_dir = "temp"
        os.makedirs(temp_dir, exist_ok=True)
        old_filename = filename
        filename = os.path.join(temp_dir, filename)

        try:
            if language == "python":
                if mut:
                    filename += 'z'
                    test_code = test_code.replace(old_filename, old_filename + 'z')
                    old_filename += 'z'
                filename_task = filename + ".py"

                with open(filename_task, 'w') as f:
                    f.write(task_code)

                filename_test = filename + "_test.py"

                with open(filename_test, 'w') as f:
                    f.write(test_code)

                original_stdout = sys.stdout
                original_stderr = sys.stderr
                loader = unittest.TestLoader()
                suite = loader.discover(start_dir='./temp', pattern=old_filename + "_test.py")

                output_capture = io.StringIO()
                # Создание test runner и запуск тестов
                with redirect_stdout(output_capture), redirect_stdout(output_capture):
                    runner = unittest.TextTestRunner(stream=output_capture)
                    result = runner.run(suite)

                sys.stdout = original_stdout
                sys.stderr = original_stderr

                return {
                    "all": result.testsRun,
                    "errors": len(result.errors), # не скомпилировался
                    "failed": len(result.failures) # не прошел тест
                }
            else:  # go
                filename_task = filename + ".go"

                with open(filename_task, 'w') as f:
                    f.write(task_code)

                filename_test = filename + "fg.go"

                with open(filename_test, 'w') as f:
                    f.write(test_code)

                result = subprocess.run(
                    ['go', 'test', '-v', './temp'],
                    text=True,
                    capture_output=True
                )

                output = result.stdout
                passed = re.findall(r'--- PASS: (\w+)', output)
                failed = re.findall(r'--- FAIL: (\w+)', output)

                return {
                    "all": len(passed),
                    "errors": 0, # не скомпилировался
                    "failed": len(failed) # не прошел тест
                }
        except subprocess.TimeoutExpired:
            return {"success": False, "output": "Timeout expired"}
        except Exception as e:
            return {"success": False, "output": str(e)}
        finally:
            # Удаляем временные файлы
            try:
                if os.path.exists(filename_task):
                    os.remove(filename_task)
                if os.path.exists(filename_test):
                    os.remove(filename_test)
                if language == "go":
                    go_mod = os.path.join(temp_dir, "go.mod")
                    if os.path.exists(go_mod):
                        os.remove(go_mod)
                    if os.path.exists("./program"):
                        os.remove("./program")

            except:
                pass

    def _print_final_results(self):
        print("\nFinal Results:")
        print("==============")

        # Unit tests results
        if self.results["unit"]["total"] > 0:
            executable_rate = self.results["unit"]["executable"] / self.results["unit"]["total"]
            mutated_rate = self.results["unit"]["mutated"] / self.results["unit"]["total"]
            spam_rate = self.results["unit"]["spamy"] / self.results["unit"]["total"]
            print(f"Unit Tests:")
            print(f"  - Executable: {executable_rate:.2%} (tests compiled and ran successfully)")
            print(f"  - Mutation Detection: {mutated_rate:.2%} (tests caught mutations)")
            print(f"  - Spam Detection: {spam_rate:.2%} (spamy tests)")

        # Integrational tests results
        if self.results["integrational"]["total_correct"] > 0:
            correct_rate = self.results["integrational"]["correct"] / self.results["integrational"]["total_correct"]
            incorrect_rate = self.results["integrational"]["incorrect"] / self.results["integrational"]["total_incorrect"]
            print(f"\nIntegrational Tests:")
            print(f"  - Correct Passed: {correct_rate:.2%} (correct tests passed)")
            print(f"  - Incorrect Failed: {incorrect_rate:.2%} (incorrect tests failed as expected)")

        # Functional tests results
        if self.results["functional"]["total"] > 0:
            passed_rate = self.results["functional"]["passed"] / self.results["functional"]["total"]
            print(f"\nFunctional Tests:")
            print(f"  - Passed: {passed_rate:.2%} (requirements verified)")

for model in MODEL_NAMES:
    benchmark = Benchmark("./Datasets", model)
    benchmark.run()