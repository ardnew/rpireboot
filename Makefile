bin = rpireboot
pkg = github.com/ardnew/$(bin)

path    = /sbin
service = $(bin).service
systemd = /etc/systemd/system

# configuration
pin      ?= 18
pull     ?= up
edge     ?= fall
debounce ?= 5ms

# executables
gpio  = gpio
logf  = printf
go    = go
cp    = cp
sudo  = sudo
ctl   = systemctl

ifeq ($(strip $(pin)),)
$(error undefined GPIO pin)
endif

ifeq ($(strip $(debounce)),)
$(error undefined debounce duration)
endif

.PHONY: all
all: gpio fmt
	@$(go) install "$(pkg)"

.PHONY: run
run: fmt
	$(go) run "$(pkg)" -p "$(pin)" -l "$(pull)" -e "$(edge)" -d "$(debounce)"

.PHONY: clean
clean: fmt
	@$(go) clean -i "$(pkg)"

.PHONY: fmt
fmt:
	@$(go) fmt "$(pkg)"

.PHONY: gpio
gpio:
	@$(logf) 'exporting GPIO%d: input, pull-up ... ' "$(pin)"
	@$(gpio) -g mode   $(pin) input
	@$(gpio) -g mode   $(pin) up
	@$(gpio)    export $(pin) input
	@$(logf) 'done\n'

.PHONY: install
install: all
	@$(sudo) $(cp) "$(shell which $(bin) )" /sbin
	@$(sudo) $(cp) "$(service)" "$(systemd)"
	@$(sudo) $(ctl) enable "$(service)"
	@$(sudo) $(ctl) start "$(service)"
