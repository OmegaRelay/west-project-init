cmake_minimum_required(VERSION 3.22)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(Zephyr REQUIRED PATHS $ENV{ZEPHYR_BASE})
project(app LANGUAGES C)

FILE(GLOB app_sources 
    src/*.c 
)

target_sources(app PRIVATE ${app_sources})
