import os
import sys
import tempfile
from pathlib import Path

import paramiko


def main() -> None:
    if len(sys.argv) != 3:
        raise SystemExit("usage: bootstrap_ssh.py HOST PUBLIC_KEY_FILE")
    password = os.environ.pop("RMPP_SSH_PASSWORD", "")
    if not password:
        raise SystemExit("RMPP_SSH_PASSWORD is required")
    host, public_key_path = sys.argv[1], Path(sys.argv[2])
    public_key = public_key_path.read_text(encoding="ascii").strip()
    if not public_key.endswith(" trmnl-remarkable-codex"):
        fields = public_key.split()
        public_key = f"{fields[0]} {fields[1]} trmnl-remarkable-codex"

    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.RejectPolicy())
    client.load_system_host_keys(str(Path.home() / ".ssh" / "known_hosts"))
    client.connect(host, username="root", password=password, look_for_keys=False,
                   allow_agent=False, timeout=10, auth_timeout=10)
    with tempfile.NamedTemporaryFile("w", encoding="ascii", delete=False) as temp:
        temp.write(public_key + "\n")
        local_temp = temp.name
    remote_temp = "/home/root/.ssh/.trmnl-key.tmp"
    sftp = client.open_sftp()
    try:
        sftp.put(local_temp, remote_temp)
        sftp.chmod(remote_temp, 0o600)
    finally:
        sftp.close()
        os.unlink(local_temp)
    command = """
set -eu
umask 077
mkdir -p /home/root/.ssh /home/root/.local/share/trmnl-remarkable/install-backup
auth=/home/root/.ssh/authorized_keys
backup=/home/root/.local/share/trmnl-remarkable/install-backup/authorized_keys.pre-trmnl
if [ -e "$auth" ] && [ ! -e "$backup" ]; then cp -p "$auth" "$backup"; fi
touch "$auth"
key=$(cat /home/root/.ssh/.trmnl-key.tmp)
if ! grep -Fqx "$key" "$auth"; then printf '%s\n' "$key" >> "$auth"; fi
rm -f /home/root/.ssh/.trmnl-key.tmp
chmod 700 /home/root/.ssh
chmod 600 "$auth"
printf 'KEY_AUTHORIZED\n'
"""
    _, stdout, stderr = client.exec_command(command, timeout=15)
    result = stdout.read().decode().strip()
    error = stderr.read().decode().strip()
    status = stdout.channel.recv_exit_status()
    client.close()
    if status != 0:
        raise SystemExit(f"bootstrap failed: {error}")
    print(result)


if __name__ == "__main__":
    main()
