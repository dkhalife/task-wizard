package app.taskwiz.di

import android.content.Context
import android.content.SharedPreferences
import app.taskwiz.auth.AuthManager
import app.taskwiz.auth.AuthTokenProvider
import app.taskwiz.data.AppPreferences
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
    fun provideLocalIdGenerator(prefs: android.content.SharedPreferences): app.taskwiz.data.LocalIdGenerator =
        app.taskwiz.data.LocalIdGenerator(prefs)

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
