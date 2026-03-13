// This is a simple calculator class
class Calculator {
    fun add(a: Double, b: Double): Double {
        return a + b
    }

    fun subtract(a: Double, b: Double): Double {
        return a - b
    }

    fun multiply(a: Double, b: Double): Double {
        return a * b
    }

    fun divide(a: Double, b: Double): Double {
        if (b == 0.0) throw IllegalArgumentException("Cannot divide by zero")
        return a / b
    }

    fun squareRoot(a: Double): Double {
        if (a < 0.0) throw IllegalArgumentException("Cannot calculate square root of negative number")
        return kotlin.math.sqrt(a)
    }
}