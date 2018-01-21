//+build windows

package cmd

/*
#include <windows.h>
void hideWindow() {
  ShowWindow(GetConsoleWindow(), SW_HIDE);
}
*/
import "C"

func hideConsoleWindow() {
	C.hideWindow()
}
