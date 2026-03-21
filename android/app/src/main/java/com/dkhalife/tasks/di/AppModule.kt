package com.dkhalife.tasks.di

import android.content.Context
import android.content.SharedPreferences
import com.dkhalife.tasks.auth.AuthManager
import com.dkhalife.tasks.auth.AuthTokenProvider
import com.dkhalife.tasks.data.AppPreferences
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object AppModule {

    @Provides
    @Singleton
    fun provideSharedPreferences(@ApplicationContext context: Context): SharedPreferences {
        return context.getSharedPreferences(AppPreferences.PREFS_NAME, Context.MODE_PRIVATE)
    }
}

@Module
@InstallIn(SingletonComponent::class)
abstract class AppBindingsModule {

    @Binds
    @Singleton
    abstract fun bindAuthTokenProvider(authManager: AuthManager): AuthTokenProvider
}
