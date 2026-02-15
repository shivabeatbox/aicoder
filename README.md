# Calculator App

A simple calculator application that supports basic arithmetic operations.

## Features

- Addition
- Subtraction
- Multiplication
- Division (with zero-division protection)

## Usage

```python
from calculator import Calculator

calc = Calculator()

# Addition
result = calc.add(5, 3)  # Returns 8

# Subtraction
result = calc.subtract(5, 3)  # Returns 2

# Multiplication
result = calc.multiply(5, 3)  # Returns 15

# Division
result = calc.divide(6, 3)  # Returns 2.0
```

## Running the Application

```bash
python calculator.py
```

## Running Tests

```bash
python test_calculator.py
```

or with unittest discovery:

```bash
python -m unittest discover
```
