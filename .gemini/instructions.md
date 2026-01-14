
## Global Programming Instructions & Guidelines

- **Code removal:** Do not remove code that is not immediately relevant 
  to the changes you want to operate when being given a specific task.
- **Replacements:** When having to do many small (and rather or completely 
  identical) replacements in one or more files, perform these replacements 
  in a single call for each file, whenever possible.

## Python Programming Guidelines

Here are some guidelines that I find useful when working on a Python codebase.

### General Guidelines

*   **Follow Project Conventions**: Adhere to the existing coding style, patterns, and practices used in the project.
*   **Dependency Management**: Use the project's dependency manager (`uv`).

### Style and Formatting

*   **PEP 8**: Follow the [PEP 8](https://www.python.org/dev/peps/pep-0008/) style guide for Python code. Use tools like `black` for automatic formatting and `ruff` for linting to enforce it.
*   **Docstrings**: Write clear and concise docstrings for all modules, functions, classes, and methods, following the [PEP 257](https://www.python.org/dev/peps/pep-0257/) conventions. Use a consistent format like Google Style or reStructuredText.
*   **Typing**: Use Python's type hints (`str`, `int`, `List`, `Dict`) for all function signatures and variables where it improves clarity. Use a static type checker like `mypy` to verify type correctness.
*   **Naming**:
    *   `snake_case` for functions, methods, and variables.
    *   `PascalCase` for classes.
    *   `UPPER_SNAKE_CASE` for constants.
    *   `_` for unused variables.

### Code Organization

*   **Modularity**: Break down large files into smaller, more manageable modules with a single responsibility.
*   **Imports**:
    *   Import modules, not individual functions or classes, to avoid circular dependencies and naming conflicts (e.g., `import my_module` instead of `from my_module import my_function`).
    *   Group imports in the following order:
        1.  Standard library imports (`os`, `sys`).
        2.  Third-party library imports (`requests`, `pandas`).
        3.  Local application/library specific imports.
*   **Absolute vs. Relative Imports**: Prefer absolute imports (`from my_app.core import utils`) over relative imports (`from ..core import utils`) for clarity and to avoid ambiguity.

### Testing

*   **Unit Tests**: Write unit tests for all new code. Use a testing framework like `pytest`.
*   **Test Coverage**: Aim for high test coverage, but focus on testing the logic and edge cases rather than just the line count.
*   **Test Naming**: Name test files `test_*.py` and test functions `test_*()`.
*   **Mocking**: Use mocking libraries like `unittest.mock` to isolate units of code and avoid external dependencies in tests.

### Documentation

*   **Comments**: Add comments to explain *why* something is done, not *what* is being done. The code itself should be self-explanatory.
*   **Configuration**: Keep configuration separate from code. Use environment variables or configuration files (`.env`, `config.ini`, `settings.py`).

### Best Practices

*   **List Comprehensions**: Use list comprehensions for creating lists in a concise and readable way.
*   **Generators**: Use generators and generator expressions for memory-efficient iteration over large datasets.
*   **Error Handling**: Be specific in exception handling. Avoid bare `except:` clauses.
*   **Context Managers**: Use the `with` statement when working with resources that need to be cleaned up (e.g., files, database connections).
