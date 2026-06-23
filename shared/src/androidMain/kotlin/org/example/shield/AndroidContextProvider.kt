package org.example.shield

import android.annotation.SuppressLint
import android.content.Context

@SuppressLint("StaticFieldLeak")
object AndroidContextProvider {
    lateinit var context: Context
}
