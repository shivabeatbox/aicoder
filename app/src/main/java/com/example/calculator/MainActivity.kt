package com.example.calculator

import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class MainActivity : AppCompatActivity() {
    private lateinit var number1: EditText
    private lateinit var number2: EditText
    private lateinit var result: TextView
    private lateinit var addButton: Button
    private lateinit var subtractButton: Button
    private lateinit var multiplyButton: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        number1 = findViewById(R.id.number1)
        number2 = findViewById(R.id.number2)
        result = findViewById(R.id.result)
        addButton = findViewById(R.id.addButton)
        subtractButton = findViewById(R.id.subtractButton)
        multiplyButton = findViewById(R.id.multiplyButton)

        addButton.setOnClickListener {
            performOperation { a, b -> a + b }
        }

        subtractButton.setOnClickListener {
            performOperation { a, b -> a - b }
        }

        multiplyButton.setOnClickListener {
            performOperation { a, b -> a * b }
        }
    }

    private fun performOperation(operation: (Double, Double) -> Double) {
        val num1 = number1.text.toString().toDoubleOrNull()
        val num2 = number2.text.toString().toDoubleOrNull()

        if (num1 != null && num2 != null) {
            val res = operation(num1, num2)
            result.text = "Result: $res"
        } else {
            result.text = "Please enter valid numbers"
        }
    }
}