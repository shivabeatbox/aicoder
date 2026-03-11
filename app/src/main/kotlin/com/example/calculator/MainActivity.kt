// MainActivity.kt

package com.example.calculator

import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.util.Log

class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        // Print hello1234 as per SAM1-29
        println("hello1234")
        Log.d("MainActivity", "hello1234")
    }
}