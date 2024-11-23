#include "common.hpp"
#include <cstring>
#include <map>
#include <set>
#include <optional>
#include <chrono>
#include <future>

namespace {

constexpr uint32_t UDP_PORT = 8888;
constexpr uint32_t TCP_PORT = 8889;
constexpr size_t BUFFER_SIZE = 1024;
constexpr int DISCOVERY_TIMEOUT_MS = 1000;
constexpr int TASK_TIMEOUT_MS = 5000;

struct ServerInfo {
    std::string ip;
    bool available;
    double from;
    double to;
    bool task_sent;
    bool result_received;
};

std::map<std::string, ServerInfo> servers;
std::set<std::string> active_servers;

void DiscoverServers() {
    active_servers.clear();
    servers.clear();
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        perror("Socket creation failed");
        exit(1);
    }

    int broadcast = 1;
    if (setsockopt(sock, SOL_SOCKET, SO_BROADCAST, &broadcast, sizeof(broadcast)) < 0) {
        perror("Broadcast option failed");
        exit(1);
    }

    struct sockaddr_in broadcast_addr;
    memset(&broadcast_addr, 0, sizeof(broadcast_addr));
    broadcast_addr.sin_family = AF_INET;
    broadcast_addr.sin_port = htons(UDP_PORT);
    broadcast_addr.sin_addr.s_addr = htonl(INADDR_BROADCAST);

    char buffer[BUFFER_SIZE];
    const char* message = "ping";

    if (sendto(sock, message, strlen(message), 0,
               reinterpret_cast<struct sockaddr*>(&broadcast_addr),
               sizeof(broadcast_addr)) < 0) {
        perror("Broadcast send failed");
        exit(1);
    }

    struct timeval tv;
    tv.tv_sec = DISCOVERY_TIMEOUT_MS / 1000;
    tv.tv_usec = (DISCOVERY_TIMEOUT_MS % 1000) * 1000;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv));

    while (true) {
        struct sockaddr_in server_addr;
        socklen_t server_len = sizeof(server_addr);
        ssize_t bytes_received = recvfrom(sock, buffer, BUFFER_SIZE, 0,
                                        reinterpret_cast<struct sockaddr*>(&server_addr),
                                        &server_len);

        if (bytes_received < 0) {
            if (errno == EAGAIN || errno == EWOULDBLOCK) {
                break;
            }
            perror("Receive failed");
            continue;
        }

        if (bytes_received == 4 && strncmp(buffer, "pong", 4) == 0) {
            std::string ip = inet_ntoa(server_addr.sin_addr);
            if (servers.find(ip) == servers.end()) {
                servers[ip] = {ip, true, 0.0, 0.0, false, false};
                active_servers.insert(ip);
            }
        }
    }

    close(sock);
}

std::optional<double> SendTask(const std::string& ip, common::Task task) {
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        servers[ip].available = false;
        return std::nullopt;
    }

    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(TCP_PORT);
    inet_pton(AF_INET, ip.c_str(), &server_addr.sin_addr);

    struct timeval tv;
    tv.tv_sec = TASK_TIMEOUT_MS / 1000;
    tv.tv_usec = (TASK_TIMEOUT_MS % 1000) * 1000;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv));

    if (connect(sock, reinterpret_cast<struct sockaddr*>(&server_addr),
                sizeof(server_addr)) < 0) {
        servers[ip].available = false;
        close(sock);
        return std::nullopt;
    }

    if (send(sock, &task, sizeof(task), 0) != sizeof(task)) {
        servers[ip].available = false;
        close(sock);
        return std::nullopt;
    }

    double result;
    if (recv(sock, &result, sizeof(result), 0) != sizeof(result)) {
        servers[ip].available = false;
        close(sock);
        return std::nullopt;
    }

    close(sock);
    servers[ip].available = true;
    servers[ip].task_sent = true;
    servers[ip].result_received = true;
    return result;
}

double CalculateIntegral(double from, double to, double step) {
    std::vector<common::Task> tasks;
    double segment_size = (to - from) / active_servers.size();
    double current_from = from;

    // Create initial tasks
    while (current_from < to) {
        double current_to = std::min(current_from + segment_size, to);
        tasks.emplace_back(current_from, current_to, step);
        current_from = current_to;
    }

    double total_result = 0.0;

    while (!tasks.empty()) {
        auto server = active_servers.begin();
        std::vector<decltype(active_servers)::iterator> servers_to_remove;
        std::vector<common::Task> tasks_to_retry;
        std::vector<std::future<std::optional<double>>> futures;
        std::vector<common::Task> pending_tasks;

        for (size_t i = 0; i < active_servers.size() && i < tasks.size(); ++i) {
            auto task = std::move(tasks.back());
            tasks.pop_back();
            pending_tasks.push_back(task);
            futures.push_back(std::async(std::launch::async, SendTask, *server, task));
            server++;
        }

        server = active_servers.begin();
        size_t i = 0;
        for (auto& future : futures) {
            auto result = future.get();
            if (result) {
                total_result += *result;
            } else {
                servers_to_remove.push_back(server);
                tasks_to_retry.push_back(std::move(pending_tasks.back()));
            }
            server++;
            i++;
        }

        for (auto server : servers_to_remove) {
            active_servers.erase(server);
        }

        tasks.insert(tasks.end(), tasks_to_retry.begin(), tasks_to_retry.end());
    }

    return total_result;
}

} // anonymous namespace

int main() {
    DiscoverServers();
    printf("Active servers: %zu\n", active_servers.size());
    double result = CalculateIntegral(0.0, 10.0, 0.1);
    printf("Result: %f\n", result);
    return 0;
}
