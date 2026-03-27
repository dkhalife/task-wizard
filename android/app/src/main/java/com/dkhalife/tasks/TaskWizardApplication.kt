package com.dkhalife.tasks

import android.app.Application
import androidx.work.Configuration
import com.dkhalife.tasks.auth.AuthManager
import com.dkhalife.tasks.data.sync.TaskSyncWorkerFactory
import com.dkhalife.tasks.telemetry.TelemetryManager
import com.microsoft.identity.client.IPublicClientApplication
import com.microsoft.identity.client.ISingleAccountPublicClientApplication
import com.microsoft.identity.client.PublicClientApplication
import com.microsoft.identity.client.exception.MsalException
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

    override val workManagerConfiguration: Configuration
        get() = Configuration.Builder()
            .setWorkerFactory(taskSyncWorkerFactory)
            .build()

    override fun onCreate() {
        super.onCreate()
        telemetryManager.initialize(this)
        setupCrashHandler()
        initializeMsal()
    }

    private fun setupCrashHandler() {
        val defaultHandler = Thread.getDefaultUncaughtExceptionHandler()
        Thread.setDefaultUncaughtExceptionHandler { thread, throwable ->
            telemetryManager.trackException(throwable, mapOf("source" to "uncaught_exception"))
            defaultHandler?.uncaughtException(thread, throwable)
        }
    }

    private fun initializeMsal() {
        PublicClientApplication.createSingleAccountPublicClientApplication(
            this,
            R.raw.auth_config_single_account,
            object : IPublicClientApplication.ISingleAccountApplicationCreatedListener {
                override fun onCreated(application: ISingleAccountPublicClientApplication) {
                    authManager.registerSingleAccountApp(application)
                    loadCurrentAccount(application)
                }

                override fun onError(exception: MsalException) {
                    telemetryManager.logError(TAG, "Failed to initialize MSAL", exception)
                }
            }
        )
    }

    private fun loadCurrentAccount(app: ISingleAccountPublicClientApplication) {
        app.getCurrentAccountAsync(object : ISingleAccountPublicClientApplication.CurrentAccountCallback {
            override fun onAccountLoaded(activeAccount: com.microsoft.identity.client.IAccount?) {
                authManager.updateAccount(activeAccount)
            }

            override fun onAccountChanged(
                priorAccount: com.microsoft.identity.client.IAccount?,
                currentAccount: com.microsoft.identity.client.IAccount?
            ) {
                authManager.updateAccount(currentAccount)
            }

            override fun onError(exception: MsalException) {
                telemetryManager.logError(TAG, "Failed to load current account", exception)
            }
        })
    }

    companion object {
        private const val TAG = "TaskWizardApplication"
    }
}
