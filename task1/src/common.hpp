#pragma once

#include <cstdint>
#include <vector>
#include <sys/socket.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <thread>

namespace common {

class Task {
public:
    Task() = default;
    Task(double from, double to, double step);

    double Calculate(std::function<double(double)> func) const;

public:
    double from_;
    double to_;
    double step_;
};

} // namespace common
