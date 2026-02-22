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

    fun isPrime(n: Int): Boolean {
        if (n <= 1) return false
        if (n == 2) return true
        if (n % 2 == 0) return false

        val sqrtN = Math.sqrt(n.toDouble()).toInt()
        for (i in 3..sqrtN step 2) {
            if (n % i == 0) return false
        }
        return true
    }
}