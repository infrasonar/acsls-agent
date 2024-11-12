[![CI](https://github.com/infrasonar/acsls-agent/workflows/CI/badge.svg)](https://github.com/infrasonar/acsls-agent/actions)
[![Release Version](https://img.shields.io/github/release/infrasonar/acsls-agent)](https://github.com/infrasonar/acsls-agent/releases)

# InfraSonar Automated Cartridge System Library Software (ACSLS) Agent

Documentation: https://docs.infrasonar.com/collectors/agents/acsls/

## Environment variables

Environment                 | Default                               | Description
----------------------------|---------------------------------------|-------------------
`CONFIG_PATH`       		| `/etc/infrasonar` 			        | Path where configuration files are loaded and stored _(note: for a user, the `$HOME` path will be used instead of `/etc`)_
`TOKEN`                     | _required_                            | Token used for authentication _(This MUST be a container token)_.
`ASSET_NAME`                | _none_                                | Initial Asset Name. This will only be used at the announce. Once the asset is created, `ASSET_NAME` will be ignored.
`ASSET_ID`                  | _none_                                | Asset Id _(If not given, the asset Id will be stored and loaded from file)_.
`API_URI`                   | https://api.infrasonar.com            | InfraSonar API.
`SKIP_VERIFY`				| _none_						        | Set to `1` or something else to skip certificate validation.
`ACSSS_STATUS_EXEC`         | `acsss status`                        | Execute acsss status, for example: `/export/home/ACSSS/bin/acsss status`.
`LIB_CMD_PHYSICAL_EXEC`     | `lib_cmd display library physical all`| Execute lib_cmd output, for example: `/export/home/ACSSS/bin/lib_cmd display library physical all`.
`CHECK_LIB_CMD_INTERVAL`    | `300`                                 | Interval in seconds for the `lib_cmd` check or `0` to disable the check.
`CHECK_ACSSS_INTERVAL`      | `300`                                 | Interval in seconds for the `acsss` check or `0` to disable the check.


## Build
```
CGO_ENABLED=0 go build -trimpath -o acsls-agent
```

## Installation

Download the latest release:
```bash
wget https://github.com/infrasonar/acsls-agent/releases/download/v0.1.0/acsls-agent
```

> _The pre-build binary is build for the **acsls-amd64** platform. For other platforms build from source using the command:_ `CGO_ENABLED=0 go build -o acsls-agent`

Ensure the binary is executable:
```
chmod +x acsls-agent
```

Copy the binary to `/usr/sbin/infrasonar-acsls-agent`

```
sudo cp acsls-agent /usr/sbin/infrasonar-acsls-agent
```

### Using Systemd

```bash
sudo touch /etc/systemd/system/infrasonar-acsls-agent.service
sudo chmod 664 /etc/systemd/system/infrasonar-acsls-agent.service
```

**1. Using you favorite editor, add the content below to the file created:**

```
[Unit]
Description=InfraSonar ACSLS Agent
Wants=network.target

[Service]
EnvironmentFile=/etc/infrasonar/acsls-agent.env
ExecStart=/usr/sbin/infrasonar-acsls-agent

[Install]
WantedBy=multi-user.target
```

**2. Create the directory `/etc/infrasonar`**

```bash
sudo mkdir /etc/infrasonar
```

**3. Create the file `/etc/infrasonar/acsls-agent.env` with at least:**

```
TOKEN=<YOUR TOKEN HERE>
```

Optionaly, add environment variable to the `acsls-agent.env` file for settings like `ASSET_ID` or `CONFIG_PATH` _(see all [environment variables](#environment-variables) in the table above)_.

**4. Reload systemd:**

```bash
sudo systemctl daemon-reload
```

**5. Install the service:**

```bash
sudo systemctl enable infrasonar-acsls-agent
```

**Finally, you may want to start/stop or view the status:**
```bash
sudo systemctl start infrasonar-acsls-agent
sudo systemctl stop infrasonar-acsls-agent
sudo systemctl status infrasonar-acsls-agent
```

**View logging:**
```bash
journalctl -u infrasonar-acsls-agent
```
