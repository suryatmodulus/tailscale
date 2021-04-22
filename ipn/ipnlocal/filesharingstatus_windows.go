// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build windows

package ipnlocal

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	sendFileShellKey = "*\\shell\\tailscale"
)

var (
	ipnExePath = ""
)

// Should be called from a locked code
func setFileSharingEnabled(enabled bool) error {
	if !enabled {
		return DisableFileSharing()
	}

	if ipnExePath == "" {
		// Find the absolute path of tailscale-ipn.exe assuming that it's in the same
		// directory as this executable (tailscaled.exe).
		var err error
		if ipnExePath, err = os.Executable(); err != nil {
			return fmt.Errorf("os.Executable error: %v", err)
		}
		if ipnExePath, err = filepath.EvalSymlinks(ipnExePath); err != nil {
			ipnExePath = ""
			return fmt.Errorf("filepath.EvalSymlinks error: %v", err)
		}
		ipnExePath = filepath.Join(filepath.Dir(ipnExePath), "tailscale-ipn.exe")
		ipnExePath, err = filepath.Abs(ipnExePath)
		if err != nil {
			ipnExePath = ""
			return fmt.Errorf("filepath.Abs error: %v", err)
		}
	}

	k, _, err := registry.CreateKey(registry.CLASSES_ROOT, sendFileShellKey, registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to create HKEY_CLASSES_ROOT\\%s reg key: %v", sendFileShellKey, err)
	}
	defer k.Close()
	if err := k.SetStringValue("", "Send with Tailscale..."); err != nil {
		return fmt.Errorf("Key.SetStringValue error: %v", err)
	}
	if err := k.SetStringValue("Icon", ipnExePath+",0"); err != nil {
		return fmt.Errorf("Key.SetStringValue error: %v", err)
	}
	c, _, err := registry.CreateKey(k, "command", registry.WRITE)
	if err != nil {
		return fmt.Errorf("failed to create HKEY_CLASSES_ROOT\\%s\\command reg key: %v", sendFileShellKey, err)
	}
	defer c.Close()
	if err := c.SetStringValue("", "\""+ipnExePath+"\" /push \"%1\""); err != nil {
		return fmt.Errorf("Key.SetStringValue error: %v", err)
	}
	return nil
}

// DisableFileSharing removes file sharing from Windows shell.
func DisableFileSharing() error {
	if err := registry.DeleteKey(registry.CLASSES_ROOT, sendFileShellKey+"\\command"); err != nil &&
		err != registry.ErrNotExist {
		return fmt.Errorf("registry.DeleteKey error: %v\n", err)
	}
	if err := registry.DeleteKey(registry.CLASSES_ROOT, sendFileShellKey); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("registry.DeleteKey error: %v\n", err)
	}
	return nil
}
