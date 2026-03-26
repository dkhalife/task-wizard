package com.dkhalife.tasks.data.calendar

import android.app.Service
import android.content.Intent
import android.os.IBinder

class CalendarAuthenticatorService : Service() {
    private lateinit var authenticator: CalendarAccountAuthenticator

    override fun onCreate() {
        super.onCreate()
        authenticator = CalendarAccountAuthenticator(this)
    }

    override fun onBind(intent: Intent?): IBinder = authenticator.iBinder
}
