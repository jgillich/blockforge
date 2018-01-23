cmake_minimum_required(VERSION 2.8.0)

project(hash C)

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -maes")
#set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -maes -msse2 -Ofast")

add_subdirectory(cryptonight)

FILE(GLOB SOURCES *.c)

add_library(hash STATIC $<TARGET_OBJECTS:cryptonight> ${SOURCES})
