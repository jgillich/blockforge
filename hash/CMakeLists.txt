cmake_minimum_required(VERSION 2.8.0)

project(hash C)

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -mtune=generic -maes -msse2 -Ofast")

add_subdirectory(cryptonight)
add_subdirectory(ethash)

file(GLOB SOURCES *.c)

add_library(hash STATIC $<TARGET_OBJECTS:cryptonight> $<TARGET_OBJECTS:ethash> ${SOURCES})

install(TARGETS hash DESTINATION "lib")
