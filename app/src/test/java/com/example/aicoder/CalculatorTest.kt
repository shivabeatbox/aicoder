package com.example.aicoder

import org.junit.Assert.assertEquals
import org.junit.Test

class CalculatorTest {
    private val calculator = Calculator()

    @Test
    fun testMultiply() {
        assertEquals(6, calculator.multiply(2, 3))
        assertEquals(0, calculator.multiply(5, 0))
        assertEquals(-15, calculator.multiply(3, -5))
        assertEquals(100, calculator.multiply(10, 10))
    }

    @Test
    fun testAdd() {
        assertEquals(5, calculator.add(2, 3))
    }

    @Test
    fun testSubtract() {
        assertEquals(1, calculator.subtract(3, 2))
    }
}