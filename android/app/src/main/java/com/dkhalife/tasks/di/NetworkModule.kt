package com.dkhalife.tasks.di

import com.dkhalife.tasks.api.ApiEndpointProvider
import com.dkhalife.tasks.api.AuthInterceptor
import com.dkhalife.tasks.api.TaskWizardApi
import com.dkhalife.tasks.auth.AuthTokenProvider
import com.google.gson.Gson
import com.google.gson.GsonBuilder
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideGson(): Gson = GsonBuilder().create()

    @Provides
    @Singleton
    fun provideOkHttpClient(tokenProvider: AuthTokenProvider): OkHttpClient {
        return OkHttpClient.Builder()
            .addInterceptor(AuthInterceptor(tokenProvider))
            .addInterceptor(HttpLoggingInterceptor().apply {
                level = HttpLoggingInterceptor.Level.BODY
            })
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(
        client: OkHttpClient,
        gson: Gson,
        endpointProvider: ApiEndpointProvider
    ): Retrofit {
        return Retrofit.Builder()
            .baseUrl(endpointProvider.getBaseUrl() + "/")
            .client(client)
            .addConverterFactory(GsonConverterFactory.create(gson))
            .build()
    }

    @Provides
    @Singleton
    fun provideTaskWizardApi(retrofit: Retrofit): TaskWizardApi {
        return retrofit.create(TaskWizardApi::class.java)
    }
}
