//go:build ignore
#include <stdlib.h>
#define APP "Trayrunner.exe "   // app to launch
#define ARGS "goPostPro "       // arguments

int main() {
    char command[] = "start /B " APP ARGS "> nul 2>&1";
    // printf("%s", command);
    system(command);

    return 0;
}
