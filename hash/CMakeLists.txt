cmake_minimum_required(VERSION 3.0)

project(hash)

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -maes")

add_subdirectory(cryptonight)

FILE(GLOB SOURCES *.c)

add_library(hash STATIC $<TARGET_OBJECTS:cryptonight> ${SOURCES})
