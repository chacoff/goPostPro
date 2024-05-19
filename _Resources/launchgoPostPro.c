//go:build ignore
#include <stdlib.h>

int main() {
    char command[] = "start /B Trayrunner.exe goPostPro > nul 2>&1""start /B Trayrunner.exe goPostPro > nul 2>&1"
    system(command);

    return 0;
}
