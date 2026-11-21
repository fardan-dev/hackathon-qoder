package com.revas.mobile.android.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val LightColors = lightColorScheme(
    primary = Color(0xFF005A9C),
    onPrimary = Color.White,
    secondary = Color(0xFFE65100),
    onSecondary = Color.White,
    error = Color(0xFFB00020),
    surface = Color(0xFFFBFCFE),
    background = Color(0xFFF5F7FA),
)

@Composable
fun RevasTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = LightColors,
        content = content,
    )
}
