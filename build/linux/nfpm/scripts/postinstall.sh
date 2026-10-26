#!/bin/sh

# The application menu learns about the new .desktop file.
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database -q /usr/share/applications
fi

exit 0
