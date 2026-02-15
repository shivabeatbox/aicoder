package com.example.aicoder;

import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import androidx.appcompat.app.AppCompatActivity;

public class MainActivity extends AppCompatActivity {

    private EditText etNumber;
    private EditText etPercentage;
    private Button btnCalculate;
    private TextView tvResult;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        etNumber = findViewById(R.id.etNumber);
        etPercentage = findViewById(R.id.etPercentage);
        btnCalculate = findViewById(R.id.btnCalculate);
        tvResult = findViewById(R.id.tvResult);

        btnCalculate.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                calculatePercentage();
            }
        });
    }

    private void calculatePercentage() {
        try {
            String numberStr = etNumber.getText().toString();
            String percentageStr = etPercentage.getText().toString();

            if (numberStr.isEmpty() || percentageStr.isEmpty()) {
                tvResult.setText("Please enter both values");
                return;
            }

            double number = Double.parseDouble(numberStr);
            double percentage = Double.parseDouble(percentageStr);

            double result = (number * percentage) / 100;

            tvResult.setText(String.format("Result: %.2f", result));
        } catch (NumberFormatException e) {
            tvResult.setText("Invalid input. Please enter valid numbers.");
        }
    }
}