//go:build ignore
#include <stdlib.h>

int main() {
    // Replace Trayrunner.exe with argument goPostPro
    system("start /B Trayrunner.exe goPostPro > nul 2>&1");

    return 0;
}
