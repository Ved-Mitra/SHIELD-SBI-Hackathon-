package com.shield.shared.data

interface DataRepository {
    fun getData(): List<String>
}

class FakeMyModelRepository : DataRepository {
    override fun getData(): List<String> = emptyList()
}
