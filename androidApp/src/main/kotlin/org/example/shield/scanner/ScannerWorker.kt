package org.example.shield.scanner

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.work.CoroutineWorker
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import org.example.shield.scanner.ApkScanner
import org.example.shield.MainActivity
import org.example.shield.R
import java.util.concurrent.TimeUnit

class ScannerWorker(
    private val context: Context,
    workerParams: WorkerParameters
) : CoroutineWorker(context, workerParams) {

    override suspend fun doWork(): Result {
        val scanner = ApkScanner(context)
        val apps = scanner.scanInstalledApps(includeSyntheticDemoThreats = true)
        val threats = apps.filter { it.isThreat }

        if (threats.isNotEmpty()) {
            sendThreatNotification(threats.size, threats.first().appName)
        }

        return Result.success()
    }

    private fun sendThreatNotification(threatCount: Int, firstThreatName: String) {
        val channelId = "shield_scanner_threats"
        val notificationManager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                channelId,
                "SHIELD Threat Alerts",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Notifications for detected sideloaded banking threats and fake APKs"
            }
            notificationManager.createNotificationChannel(channel)
        }

        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        val pendingIntent = PendingIntent.getActivity(
            context,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val notification = NotificationCompat.Builder(context, channelId)
            .setSmallIcon(R.drawable.ic_stat_shield)
            .setContentTitle("SHIELD Security Alert: $threatCount Threat(s) Detected")
            .setContentText("High-risk sideloaded app detected: $firstThreatName. Tap to view details and remove.")
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setContentIntent(pendingIntent)
            .setAutoCancel(true)
            .build()

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (context.checkSelfPermission(android.Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) {
                NotificationManagerCompat.from(context).notify(1001, notification)
            }
        } else {
            NotificationManagerCompat.from(context).notify(1001, notification)
        }
    }

    companion object {
        const val WORK_NAME = "ShieldPassiveScannerWork"

        fun schedule(context: Context) {
            val workRequest = PeriodicWorkRequestBuilder<ScannerWorker>(15, TimeUnit.MINUTES)
                .build()

            WorkManager.getInstance(context).enqueueUniquePeriodicWork(
                WORK_NAME,
                ExistingPeriodicWorkPolicy.UPDATE,
                workRequest
            )
        }

        fun cancel(context: Context) {
            WorkManager.getInstance(context).cancelUniqueWork(WORK_NAME)
        }
    }
}
