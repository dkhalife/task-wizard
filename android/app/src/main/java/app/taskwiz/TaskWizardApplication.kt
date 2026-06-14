package app.taskwiz

import android.app.Application
import androidx.work.Configuration
import app.taskwiz.auth.AuthManager
import app.taskwiz.data.network.NetworkMonitor
import app.taskwiz.data.sync.SyncCoordinator
import app.taskwiz.data.sync.TaskSyncWorkerFactory
import app.taskwiz.data.sync.WebSocketLifecycleManager
import app.taskwiz.telemetry.TelemetryManager
import dagger.hilt.android.HiltAndroidApp
import javax.inject.Inject

@HiltAndroidApp
class TaskWizardApplication : Application(), Configuration.Provider {

    @Inject
    lateinit var authManager: AuthManager

    @Inject
    lateinit var taskSyncWorkerFactory: TaskSyncWorkerFactory

    @Inject
    lateinit var telemetryManager: TelemetryManager

    @Inject
    lateinit var webSocketLifecycleManager: WebSocketLifecycleManager

    @Inject
    lateinit var syncCoordinator: SyncCoordinator

    @Inject
    lateinit var networkMonitor: NetworkMonitor

    override val workManagerConfiguration: Configuration
        get() = Configuration.Builder()
            .setWorkerFactory(taskSyncWorkerFactory)
            .build()

    override fun onCreate() {
        super.onCreate()
        telemetryManager.initialize(this)
        setupCrashHandler()
        authManager.tryRestoreMsal(this)
        webSocketLifecycleManager.start()
        networkMonitor.addOnAvailableListener { syncCoordinator.syncOnce() }
        if (networkMonitor.isOnline.value) {
            syncCoordinator.syncOnce()
        }
    }

    private fun setupCrashHandler() {
        val defaultHandler = Thread.getDefaultUncaughtExceptionHandler()
        Thread.setDefaultUncaughtExceptionHandler { thread, throwable ->
            try {
                telemetryManager.trackException(throwable, mapOf("source" to "uncaught_exception"))
                try {
                    Thread.sleep(2000)
                } catch (_: InterruptedException) {
                    Thread.currentThread().interrupt()
                }
            } catch (_: Exception) {
            } finally {
                defaultHandler?.uncaughtException(thread, throwable)
            }
        }
    }

    companion object {
        private const val TAG = "TaskWizardApplication"
    }
}
