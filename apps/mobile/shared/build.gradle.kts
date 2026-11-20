import org.jetbrains.kotlin.gradle.dsl.JvmTarget
import org.jetbrains.kotlin.gradle.plugin.mpp.KotlinNativeTarget

plugins {
    alias(libs.plugins.kotlin.multiplatform)
    alias(libs.plugins.android.library)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.sqldelight)
}

kotlin {
    compilerOptions {
        optIn.add("kotlin.time.ExperimentalTime")
    }

    androidTarget {
        compilerOptions {
            jvmTarget.set(JvmTarget.JVM_17)
        }
    }

    iosArm64()
    iosSimulatorArm64()

    targets.withType<KotlinNativeTarget>().configureEach {
        binaries.framework {
            baseName = "Shared"
            isStatic = true
        }
    }

    sourceSets {
        commonMain.dependencies {
            implementation(libs.kotlinx.coroutines.core)
            implementation(libs.kotlinx.serialization.json)
            implementation(libs.kotlinx.datetime)
            implementation(libs.ktor.client.core)
            implementation(libs.ktor.client.content.negotiation)
            implementation(libs.ktor.serialization.kotlinx.json)
            implementation(libs.sqldelight.runtime)
        }
        commonTest.dependencies {
            implementation(kotlin("test"))
            implementation(libs.kotlinx.coroutines.test)
            implementation(libs.ktor.client.mock)
        }
        androidMain.dependencies {
            implementation(libs.ktor.client.okhttp)
            implementation(libs.sqldelight.android.driver)
        }
        iosMain.dependencies {
            implementation(libs.ktor.client.darwin)
            implementation(libs.sqldelight.native.driver)
        }
        val androidUnitTest by getting {
            dependencies {
                implementation(libs.sqldelight.sqlite.driver)
            }
        }
    }
}

android {
    namespace = "com.revas.mobile.shared"
    compileSdk = 35

    defaultConfig {
        minSdk = 26
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

sqldelight {
    databases {
        create("MobileDatabase") {
            packageName.set("com.revas.mobile.db")
        }
    }
}

val domainDir = layout.projectDirectory.dir("src/commonMain/kotlin/com/revas/mobile/domain")
val forbiddenDomainImports = listOf(
    "io.ktor",
    "app.cash.sqldelight",
    "androidx.",
    "android.",
    "platform.Foundation",
    "platform.UIKit",
    "kotlinx.serialization",
)

val checkDomainDependencies by tasks.registering {
    description = "Enforces the mobile dependency rule: domain must stay framework-free."
    group = "verification"
    val dir = domainDir
    val forbidden = forbiddenDomainImports
    doLast {
        val violations = mutableListOf<String>()
        dir.asFile.walkTopDown().filter { it.isFile && it.extension == "kt" }.forEach { file ->
            file.readLines().forEachIndexed { index, line ->
                val trimmed = line.trim()
                if (!trimmed.startsWith("import ")) return@forEachIndexed
                val imported = trimmed.removePrefix("import ")
                if (forbidden.any { imported.startsWith(it) }) {
                    violations += "${file.relativeTo(projectDir)}:${index + 1}: $trimmed"
                }
            }
        }
        if (violations.isNotEmpty()) {
            throw GradleException(
                "Domain dependency rule violated:\n" + violations.joinToString("\n"),
            )
        }
    }
}

tasks.named("check") {
    dependsOn(checkDomainDependencies)
}
