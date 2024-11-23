#include "common.hpp"

namespace common {

Task::Task(double from, double to, double step)
    : from_(from)
    , to_(to)
    , step_(step) {}

double Task::Calculate(std::function<double(double)> func) const {
    double result = 0;
    for (double x = from_; x < to_; x += step_) {
        result += func(x);
    }
    return result * step_;
}

} // namespace common
