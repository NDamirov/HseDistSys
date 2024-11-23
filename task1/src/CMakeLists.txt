cmake_minimum_required(VERSION 3.20)
project(distributed_math)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

add_executable(server server.cpp common.cpp)
add_executable(client client.cpp common.cpp)
