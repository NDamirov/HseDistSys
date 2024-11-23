#include "common.hpp"

namespace {

uint32_t UDP_PORT = 8888;
uint32_t TCP_PORT = 8889;
constexpr size_t BUFFER_SIZE = 1024;

void HandleClient(int client_socket, const struct sockaddr_in& client_addr) {
    common::Task task;
    char buffer[BUFFER_SIZE];
    ssize_t bytes_received = recv(client_socket, buffer, sizeof(buffer), 0);
    if (bytes_received == sizeof(task)) {
        printf("Received task from %s:%d\n", inet_ntoa(client_addr.sin_addr), ntohs(client_addr.sin_port));
        memcpy(&task, buffer, sizeof(task));
        printf("Task: from %f to %f with step %f\n", task.from_, task.to_, task.step_);
        double result = task.Calculate([](double x) { return x * x; });
        send(client_socket, &result, sizeof(result), 0);
    }

    close(client_socket);
}

void ClientDiscovery() {
    int udp_sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (udp_sock < 0) {
        perror("UDP socket creation failed");
        exit(1);
    }

    struct sockaddr_in udp_addr;
    memset(&udp_addr, 0, sizeof(udp_addr));
    udp_addr.sin_family = AF_INET;
    udp_addr.sin_addr.s_addr = INADDR_ANY;
    udp_addr.sin_port = htons(UDP_PORT);

    if (bind(udp_sock, reinterpret_cast<struct sockaddr*>(&udp_addr), sizeof(udp_addr)) < 0) {
        perror("UDP bind failed");
        exit(1);
    }

    printf("Server is running on port %d\n", UDP_PORT);

    char buffer[BUFFER_SIZE];
    while (1) {
        struct sockaddr_in client_addr;
        socklen_t client_len = sizeof(client_addr);
        ssize_t bytes_received = recvfrom(udp_sock, buffer, BUFFER_SIZE, 0,
                                        reinterpret_cast<struct sockaddr*>(&client_addr), &client_len);
        if (bytes_received > 0 && strncmp(buffer, "ping", 4) == 0) {
            sendto(udp_sock, "pong", 4, 0,
                   reinterpret_cast<struct sockaddr*>(&client_addr), client_len);
        }
    }
}

void TaskHandler() {
    int tcp_sock = socket(AF_INET, SOCK_STREAM, 0);
    if (tcp_sock < 0) {
        perror("TCP socket creation failed");
        exit(1);
    }

    struct sockaddr_in tcp_addr;
    memset(&tcp_addr, 0, sizeof(tcp_addr));
    tcp_addr.sin_family = AF_INET;
    tcp_addr.sin_addr.s_addr = INADDR_ANY;
    tcp_addr.sin_port = htons(TCP_PORT);

    if (bind(tcp_sock, reinterpret_cast<struct sockaddr*>(&tcp_addr), sizeof(tcp_addr)) < 0) {
        perror("TCP bind failed");
        exit(1);
    }

    if (listen(tcp_sock, 100) < 0) {
        perror("TCP listen failed");
        exit(1);
    }

    printf("Server is running on port %d\n", TCP_PORT);

    while (1) {
        struct sockaddr_in client_addr;
        socklen_t client_len = sizeof(client_addr);
        int client_socket = accept(tcp_sock,
                                    reinterpret_cast<struct sockaddr*>(&client_addr), &client_len);
        if (client_socket < 0) {
            perror("TCP accept failed");
            continue;
        }

        HandleClient(client_socket, client_addr);
    }
}

void Run() {
    std::thread discovery_thread(ClientDiscovery);
    std::thread task_handler_thread(TaskHandler);

    discovery_thread.join();
    task_handler_thread.join();
}

} // anonymous namespace

int main(int argc, char* argv[]) {
    if (argc != 3) {
        fprintf(stderr, "Usage: %s <udp_port> <tcp_port>\n", argv[0]);
        return 1;
    }

    UDP_PORT = static_cast<uint32_t>(std::stoi(argv[1]));
    TCP_PORT = static_cast<uint32_t>(std::stoi(argv[2]));
    Run();
    return 0;
}
