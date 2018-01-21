//+build windows

package cmd

/*
#include <windows.h>
void hideWindow() {
  ShowWindow(GetConsoleWindow(), SW_HIDE);
}
*/
import "C"
import "github.com/inconshreveable/mousetrap"

func hideConsoleWindow() {
	if mousetrap.StartedByExplorer() {
		C.hideWindow()
	}
}
