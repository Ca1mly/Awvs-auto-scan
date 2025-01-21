# 指定最低CMake版本
cmake_minimum_required(VERSION 3.10)

# 项目名称
project(MyProject)

# 添加可执行文件
add_executable(MyApp main.cpp)

# 添加库
add_library(MyLib src/lib.cpp)

# 链接库
target_link_libraries(MyApp MyLib) 