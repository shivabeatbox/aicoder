package com.example.aicoder

import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import java.text.DecimalFormat

class MainActivity : AppCompatActivity() {

    private lateinit var etNumber: EditText
    private lateinit var etPercentage: EditText
    private lateinit var btnCalculate: Button
    private lateinit var tvResult: TextView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        etNumber = findViewById(R.id.etNumber)
        etPercentage = findViewById(R.id.etPercentage)
        btnCalculate = findViewById(R.id.btnCalculate)
        tvResult = findViewById(R.id.tvResult)

        btnCalculate.setOnClickListener {
            calculatePercentage()
        }
    }

    private fun calculatePercentage() {
        val numberStr = etNumber.text.toString()
        val percentageStr = etPercentage.text.toString()

        if (numberStr.isEmpty() || percentageStr.isEmpty()) {
            tvResult.text = "Please enter both values"
            return
        }

        try {
            val number = numberStr.toDouble()
            val percentage = percentageStr.toDouble()
            val result = (number * percentage) / 100

            val df = DecimalFormat("#.##")
            tvResult.text = "Result: ${df.format(result)}"
        } catch (e: NumberFormatException) {
            tvResult.text = "Invalid input"
        }
    }
}