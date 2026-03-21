package com.dkhalife.tasks.model

data class Label(
    val id: Int = 0,
    val name: String = "",
    val color: String = "#000000",
    @com.google.gson.annotations.SerializedName("created_at") val createdAt: String? = null,
    @com.google.gson.annotations.SerializedName("updated_at") val updatedAt: String? = null
)

data class CreateLabelReq(
    val name: String,
    val color: String = "#000000"
)

data class UpdateLabelReq(
    val id: Int,
    val name: String,
    val color: String = "#000000"
)
