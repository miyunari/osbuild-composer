package rhel90

import (
	"fmt"
	"math/rand"
	"path"
	"path/filepath"
	"strings"

	"github.com/osbuild/osbuild-composer/internal/blueprint"
	"github.com/osbuild/osbuild-composer/internal/common"
	"github.com/osbuild/osbuild-composer/internal/disk"
	"github.com/osbuild/osbuild-composer/internal/distro"
	osbuild "github.com/osbuild/osbuild-composer/internal/osbuild2"
	"github.com/osbuild/osbuild-composer/internal/rpmmd"
)

func qcow2Pipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	treePipeline, err := osPipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}

	partitionTable, err := t.getPartitionTable(customizations.GetFilesystems(), options, rng)
	if err != nil {
		return nil, err
	}

	treePipeline = prependKernelCmdlineStage(treePipeline, t, &partitionTable)

	if options.Subscription == nil {
		// RHSM DNF plugins should be by default disabled on RHEL Guest KVM images
		treePipeline.AddStage(osbuild.NewRHSMStage(&osbuild.RHSMStageOptions{
			DnfPlugins: &osbuild.RHSMStageOptionsDnfPlugins{
				ProductID: &osbuild.RHSMStageOptionsDnfPlugin{
					Enabled: false,
				},
				SubscriptionManager: &osbuild.RHSMStageOptionsDnfPlugin{
					Enabled: false,
				},
			},
		}))
	}

	treePipeline.AddStage(osbuild.NewFSTabStage(partitionTable.FSTabStageOptionsV2()))
	kernelVer := kernelVerStr(packageSetSpecs[blueprintPkgsKey], customizations.GetKernel().Name, t.Arch().Name())
	treePipeline.AddStage(bootloaderConfigStage(t, partitionTable, customizations.GetKernel(), kernelVer, false, false))
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	diskfile := "disk.img"
	imagePipeline := liveImagePipeline(treePipeline.Name, diskfile, &partitionTable, t.arch, kernelVer)
	pipelines = append(pipelines, *imagePipeline)

	qemuPipeline := qemuPipeline(imagePipeline.Name, diskfile, t.filename, "qcow2", "1.1")
	pipelines = append(pipelines, *qemuPipeline)

	return pipelines, nil
}

func prependKernelCmdlineStage(pipeline *osbuild.Pipeline, t *imageType, pt *disk.PartitionTable) *osbuild.Pipeline {
	rootFsUUID := pt.RootPartition().Filesystem.UUID
	kernelStage := osbuild.NewKernelCmdlineStage(kernelCmdlineStageOptions(rootFsUUID, t.kernelOptions))
	pipeline.Stages = append([]*osbuild.Stage{kernelStage}, pipeline.Stages...)
	return pipeline
}

func vhdPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))
	treePipeline, err := osPipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}

	partitionTable, err := t.getPartitionTable(customizations.GetFilesystems(), options, rng)
	if err != nil {
		return nil, err
	}

	treePipeline = prependKernelCmdlineStage(treePipeline, t, &partitionTable)
	treePipeline.AddStage(osbuild.NewFSTabStage(partitionTable.FSTabStageOptionsV2()))
	kernelVer := kernelVerStr(packageSetSpecs[blueprintPkgsKey], customizations.GetKernel().Name, t.Arch().Name())
	treePipeline.AddStage(bootloaderConfigStage(t, partitionTable, customizations.GetKernel(), kernelVer, false, false))
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	diskfile := "disk.img"
	imagePipeline := liveImagePipeline(treePipeline.Name, diskfile, &partitionTable, t.arch, kernelVer)
	pipelines = append(pipelines, *imagePipeline)
	if err != nil {
		return nil, err
	}

	qemuPipeline := qemuPipeline(imagePipeline.Name, diskfile, t.filename, "vpc", "")
	pipelines = append(pipelines, *qemuPipeline)
	return pipelines, nil
}

func vmdkPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))
	treePipeline, err := osPipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}

	partitionTable, err := t.getPartitionTable(customizations.GetFilesystems(), options, rng)
	if err != nil {
		return nil, err
	}

	treePipeline = prependKernelCmdlineStage(treePipeline, t, &partitionTable)
	treePipeline.AddStage(osbuild.NewFSTabStage(partitionTable.FSTabStageOptionsV2()))
	kernelVer := kernelVerStr(packageSetSpecs[blueprintPkgsKey], customizations.GetKernel().Name, t.Arch().Name())
	treePipeline.AddStage(bootloaderConfigStage(t, partitionTable, customizations.GetKernel(), kernelVer, false, false))
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	diskfile := "disk.img"
	imagePipeline := liveImagePipeline(treePipeline.Name, diskfile, &partitionTable, t.arch, kernelVer)
	pipelines = append(pipelines, *imagePipeline)
	if err != nil {
		return nil, err
	}

	qemuPipeline := qemuPipeline(imagePipeline.Name, diskfile, t.filename, "vmdk", "")
	pipelines = append(pipelines, *qemuPipeline)
	return pipelines, nil
}

func openstackPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))
	treePipeline, err := osPipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}

	partitionTable, err := t.getPartitionTable(customizations.GetFilesystems(), options, rng)
	if err != nil {
		return nil, err
	}

	treePipeline = prependKernelCmdlineStage(treePipeline, t, &partitionTable)
	treePipeline.AddStage(osbuild.NewFSTabStage(partitionTable.FSTabStageOptionsV2()))
	kernelVer := kernelVerStr(packageSetSpecs[blueprintPkgsKey], customizations.GetKernel().Name, t.Arch().Name())
	treePipeline.AddStage(bootloaderConfigStage(t, partitionTable, customizations.GetKernel(), kernelVer, false, false))
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	diskfile := "disk.img"
	imagePipeline := liveImagePipeline(treePipeline.Name, diskfile, &partitionTable, t.arch, kernelVer)
	pipelines = append(pipelines, *imagePipeline)
	if err != nil {
		return nil, err
	}

	qemuPipeline := qemuPipeline(imagePipeline.Name, diskfile, t.filename, "qcow2", "")
	pipelines = append(pipelines, *qemuPipeline)
	return pipelines, nil
}

// ec2BaseTreePipeline returns the base OS pipeline common for all EC2 image types.
//
// The expectation is that specific EC2 image types can extend the returned pipeline
// by appending additional stages.
//
// The argument `withRHUI` should be set to `true` only if the image package set includes RHUI client packages.
//
// Note: the caller of this function has to append the `osbuild.NewSELinuxStage(selinuxStageOptions(false))` stage
// as the last one to the returned pipeline. The stage is not appended on purpose, to allow caller to append
// any additional stages to the pipeline, but before the SELinuxStage, which must be always the last one.
func ec2BaseTreePipeline(
	repos []rpmmd.RepoConfig,
	packages []rpmmd.PackageSpec,
	bpPackages []rpmmd.PackageSpec,
	c *blueprint.Customizations,
	options distro.ImageOptions,
	enabledServices, disabledServices []string,
	defaultTarget string,
	withRHUI, isRHEL bool,
	pt *disk.PartitionTable) (*osbuild.Pipeline, error) {

	p := new(osbuild.Pipeline)
	p.Name = "os"
	p.Build = "name:build"
	packages = append(packages, bpPackages...)
	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(packages)))

	// If the /boot is on a separate partition, the prefix for the BLS stage must be ""
	if pt.BootPartition() == nil {
		p.AddStage(osbuild.NewFixBLSStage(&osbuild.FixBLSStageOptions{}))
	} else {
		p.AddStage(osbuild.NewFixBLSStage(&osbuild.FixBLSStageOptions{Prefix: common.StringToPtr("")}))
	}

	language, keyboard := c.GetPrimaryLocale()
	if language != nil {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: *language}))
	} else {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: "en_US.UTF-8"}))
	}
	if keyboard != nil {
		p.AddStage(osbuild.NewKeymapStage(&osbuild.KeymapStageOptions{Keymap: *keyboard}))
	} else {
		p.AddStage(osbuild.NewKeymapStage(&osbuild.KeymapStageOptions{
			Keymap: "us",
			X11Keymap: &osbuild.X11KeymapOptions{
				Layouts: []string{"us"},
			},
		}))
	}

	if hostname := c.GetHostname(); hostname != nil {
		p.AddStage(osbuild.NewHostnameStage(&osbuild.HostnameStageOptions{Hostname: *hostname}))
	}

	timezone, ntpServers := c.GetTimezoneSettings()
	if timezone != nil {
		p.AddStage(osbuild.NewTimezoneStage(&osbuild.TimezoneStageOptions{Zone: *timezone}))
	} else {
		p.AddStage(osbuild.NewTimezoneStage(&osbuild.TimezoneStageOptions{Zone: "UTC"}))
	}

	if len(ntpServers) > 0 {
		p.AddStage(osbuild.NewChronyStage(&osbuild.ChronyStageOptions{Timeservers: ntpServers}))
	} else {
		p.AddStage(osbuild.NewChronyStage(&osbuild.ChronyStageOptions{
			Servers: []osbuild.ChronyConfigServer{
				{
					Hostname: "169.254.169.123",
					Prefer:   common.BoolToPtr(true),
					Iburst:   common.BoolToPtr(true),
					Minpoll:  common.IntToPtr(4),
					Maxpoll:  common.IntToPtr(4),
				},
			},
			// empty string will remove any occurrences of the option from the configuration
			LeapsecTz: common.StringToPtr(""),
		}))
	}

	if groups := c.GetGroups(); len(groups) > 0 {
		p.AddStage(osbuild.NewGroupsStage(groupStageOptions(groups)))
	}

	if users := c.GetUsers(); len(users) > 0 {
		userOptions, err := userStageOptions(users)
		if err != nil {
			return nil, err
		}
		p.AddStage(osbuild.NewUsersStage(userOptions))
	}

	if services := c.GetServices(); services != nil || enabledServices != nil || disabledServices != nil || defaultTarget != "" {
		p.AddStage(osbuild.NewSystemdStage(systemdStageOptions(enabledServices, disabledServices, services, defaultTarget)))
	}

	if firewall := c.GetFirewall(); firewall != nil {
		p.AddStage(osbuild.NewFirewallStage(firewallStageOptions(firewall)))
	}

	p.AddStage(osbuild.NewSystemdLogindStage(&osbuild.SystemdLogindStageOptions{
		Filename: "00-getty-fixes.conf",
		Config: osbuild.SystemdLogindConfigDropin{

			Login: osbuild.SystemdLogindConfigLoginSection{
				NAutoVTs: common.IntToPtr(0),
			},
		},
	}))

	p.AddStage(osbuild.NewSysconfigStage(&osbuild.SysconfigStageOptions{
		Kernel: osbuild.SysconfigKernelOptions{
			UpdateDefault: true,
			DefaultKernel: "kernel",
		},
		Network: osbuild.SysconfigNetworkOptions{
			Networking: true,
			NoZeroConf: true,
		},
		NetworkScripts: &osbuild.NetworkScriptsOptions{
			IfcfgFiles: map[string]osbuild.IfcfgFile{
				"eth0": {
					Device:    "eth0",
					Bootproto: osbuild.IfcfgBootprotoDHCP,
					OnBoot:    common.BoolToPtr(true),
					Type:      osbuild.IfcfgTypeEthernet,
					UserCtl:   common.BoolToPtr(true),
					PeerDNS:   common.BoolToPtr(true),
					IPv6Init:  common.BoolToPtr(false),
				},
			},
		},
	}))

	p.AddStage(osbuild.NewCloudInitStage(&osbuild.CloudInitStageOptions{
		Filename: "00-rhel-default-user.cfg",
		Config: osbuild.CloudInitConfigFile{
			SystemInfo: &osbuild.CloudInitConfigSystemInfo{
				DefaultUser: &osbuild.CloudInitConfigDefaultUser{
					Name: "ec2-user",
				},
			},
		},
	}))

	p.AddStage(osbuild.NewModprobeStage(&osbuild.ModprobeStageOptions{
		Filename: "blacklist-nouveau.conf",
		Commands: osbuild.ModprobeConfigCmdList{
			osbuild.NewModprobeConfigCmdBlacklist("nouveau"),
		},
	}))

	p.AddStage(osbuild.NewDracutConfStage(&osbuild.DracutConfStageOptions{
		Filename: "sgdisk.conf",
		Config: osbuild.DracutConfigFile{
			Install: []string{"sgdisk"},
		},
	}))

	// RHBZ#1822863
	p.AddStage(osbuild.NewSystemdUnitStage(&osbuild.SystemdUnitStageOptions{
		Unit:   "nm-cloud-setup.service",
		Dropin: "10-rh-enable-for-ec2.conf",
		Config: osbuild.SystemdServiceUnitDropin{
			Service: &osbuild.SystemdUnitServiceSection{
				Environment: "NM_CLOUD_SETUP_EC2=yes",
			},
		},
	}))

	p.AddStage(osbuild.NewAuthselectStage(&osbuild.AuthselectStageOptions{
		Profile: "sssd",
	}))

	if isRHEL {
		if options.Subscription != nil {
			commands := []string{
				fmt.Sprintf("/usr/sbin/subscription-manager register --org=%s --activationkey=%s --serverurl %s --baseurl %s", options.Subscription.Organization, options.Subscription.ActivationKey, options.Subscription.ServerUrl, options.Subscription.BaseUrl),
			}
			if options.Subscription.Insights {
				commands = append(commands, "/usr/bin/insights-client --register")
			}

			p.AddStage(osbuild.NewFirstBootStage(&osbuild.FirstBootStageOptions{
				Commands:       commands,
				WaitForNetwork: true,
			}))
		} else {
			// The EC2 images should keep the RHSM DNF plugins enabled (RHBZ#1996670)
			rhsmStageOptions := &osbuild.RHSMStageOptions{
				// RHBZ#1932802
				SubMan: &osbuild.RHSMStageOptionsSubMan{
					Rhsmcertd: &osbuild.SubManConfigRHSMCERTDSection{
						AutoRegistration: common.BoolToPtr(true),
					},
				},
			}

			// Disable RHSM redhat.repo management only if the image uses RHUI
			// for content. Otherwise subscribing the system manually after booting
			// it would result in empty redhat.repo. Without RHUI, such system
			// would have no way to get Red Hat content, but enable the repo
			// management manually, which would be very confusing.
			// RHBZ#1932802
			if withRHUI {
				rhsmStageOptions.SubMan.Rhsm = &osbuild.SubManConfigRHSMSection{
					ManageRepos: common.BoolToPtr(false),
				}
			}

			p.AddStage(osbuild.NewRHSMStage(rhsmStageOptions))
		}
	}

	return p, nil
}

func ec2X86_64BaseTreePipeline(repos []rpmmd.RepoConfig, packages []rpmmd.PackageSpec, bpPackages []rpmmd.PackageSpec,
	c *blueprint.Customizations, options distro.ImageOptions, enabledServices, disabledServices []string,
	defaultTarget string, withRHUI, isRHEL bool, pt *disk.PartitionTable) (*osbuild.Pipeline, error) {

	treePipeline, err := ec2BaseTreePipeline(repos, packages, bpPackages, c, options, enabledServices, disabledServices, defaultTarget, withRHUI, isRHEL, pt)
	if err != nil {
		return nil, err
	}

	// EC2 x86_64-specific stages
	// Add 'nvme' driver to handle the case when initramfs is getting forcefully
	// rebuild on a Xen instance (and not able to boot on Nitro after that).
	treePipeline.AddStage(osbuild.NewDracutConfStage(&osbuild.DracutConfStageOptions{
		Filename: "ec2.conf",
		Config: osbuild.DracutConfigFile{
			AddDrivers: []string{
				"nvme",
				"xen-blkfront",
			},
		},
	}))

	return treePipeline, nil
}

func ec2CommonPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions,
	repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec,
	rng *rand.Rand, withRHUI bool, diskfile string) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	partitionTable, err := t.getPartitionTable(customizations.GetFilesystems(), options, rng)
	if err != nil {
		return nil, err
	}

	var treePipeline *osbuild.Pipeline
	switch arch := t.arch.Name(); arch {
	// rhel-ec2-x86_64, rhel-ha-ec2
	case distro.X86_64ArchName:
		treePipeline, err = ec2X86_64BaseTreePipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget, withRHUI, t.arch.distro.isRHEL(), &partitionTable)
	// rhel-ec2-aarch64
	case distro.Aarch64ArchName:
		treePipeline, err = ec2BaseTreePipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget, withRHUI, t.arch.distro.isRHEL(), &partitionTable)
	default:
		return nil, fmt.Errorf("ec2CommonPipelines: unsupported image architecture: %q", arch)
	}
	if err != nil {
		return nil, err
	}

	treePipeline = prependKernelCmdlineStage(treePipeline, t, &partitionTable)
	treePipeline.AddStage(osbuild.NewFSTabStage(partitionTable.FSTabStageOptionsV2()))
	kernelVer := kernelVerStr(packageSetSpecs[blueprintPkgsKey], customizations.GetKernel().Name, t.Arch().Name())
	treePipeline.AddStage(bootloaderConfigStage(t, partitionTable, customizations.GetKernel(), kernelVer, false, false))
	// The last stage must be the SELinux stage
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	imagePipeline := liveImagePipeline(treePipeline.Name, diskfile, &partitionTable, t.arch, kernelVer)
	pipelines = append(pipelines, *imagePipeline)
	return pipelines, nil
}

func ec2SapPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions,
	repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec,
	rng *rand.Rand, withRHUI bool, diskfile string) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	partitionTable, err := t.getPartitionTable(customizations.GetFilesystems(), options, rng)
	if err != nil {
		return nil, err
	}

	var treePipeline *osbuild.Pipeline
	switch arch := t.arch.Name(); arch {
	// rhel-sap-ec2
	case distro.X86_64ArchName:
		treePipeline, err = ec2X86_64BaseTreePipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget, withRHUI, t.arch.distro.isRHEL(), &partitionTable)
	default:
		return nil, fmt.Errorf("ec2SapPipelines: unsupported image architecture: %q", arch)
	}
	if err != nil {
		return nil, err
	}

	// SAP-specific configuration
	treePipeline.AddStage(osbuild.NewSELinuxConfigStage(&osbuild.SELinuxConfigStageOptions{
		State: osbuild.SELinuxStatePermissive,
	}))

	// RHBZ#1960617
	treePipeline.AddStage(osbuild.NewTunedStage(osbuild.NewTunedStageOptions("sap-hana")))

	// RHBZ#1959979
	treePipeline.AddStage(osbuild.NewTmpfilesdStage(osbuild.NewTmpfilesdStageOptions("sap.conf",
		[]osbuild.TmpfilesdConfigLine{
			{
				Type: "x",
				Path: "/tmp/.sap*",
			},
			{
				Type: "x",
				Path: "/tmp/.hdb*lock",
			},
			{
				Type: "x",
				Path: "/tmp/.trex*lock",
			},
		},
	)))

	// RHBZ#1959963
	treePipeline.AddStage(osbuild.NewPamLimitsConfStage(osbuild.NewPamLimitsConfStageOptions("99-sap.conf",
		[]osbuild.PamLimitsConfigLine{
			{
				Domain: "@sapsys",
				Type:   osbuild.PamLimitsTypeHard,
				Item:   osbuild.PamLimitsItemNofile,
				Value:  osbuild.PamLimitsValueInt(65536),
			},
			{
				Domain: "@sapsys",
				Type:   osbuild.PamLimitsTypeSoft,
				Item:   osbuild.PamLimitsItemNofile,
				Value:  osbuild.PamLimitsValueInt(65536),
			},
			{
				Domain: "@dba",
				Type:   osbuild.PamLimitsTypeHard,
				Item:   osbuild.PamLimitsItemNofile,
				Value:  osbuild.PamLimitsValueInt(65536),
			},
			{
				Domain: "@dba",
				Type:   osbuild.PamLimitsTypeSoft,
				Item:   osbuild.PamLimitsItemNofile,
				Value:  osbuild.PamLimitsValueInt(65536),
			},
			{
				Domain: "@sapsys",
				Type:   osbuild.PamLimitsTypeHard,
				Item:   osbuild.PamLimitsItemNproc,
				Value:  osbuild.PamLimitsValueUnlimited,
			},
			{
				Domain: "@sapsys",
				Type:   osbuild.PamLimitsTypeSoft,
				Item:   osbuild.PamLimitsItemNproc,
				Value:  osbuild.PamLimitsValueUnlimited,
			},
			{
				Domain: "@dba",
				Type:   osbuild.PamLimitsTypeHard,
				Item:   osbuild.PamLimitsItemNproc,
				Value:  osbuild.PamLimitsValueUnlimited,
			},
			{
				Domain: "@dba",
				Type:   osbuild.PamLimitsTypeSoft,
				Item:   osbuild.PamLimitsItemNproc,
				Value:  osbuild.PamLimitsValueUnlimited,
			},
		},
	)))

	// RHBZ#1959962
	treePipeline.AddStage(osbuild.NewSysctldStage(osbuild.NewSysctldStageOptions("sap.conf",
		[]osbuild.SysctldConfigLine{
			{
				Key:   "kernel.pid_max",
				Value: "4194304",
			},
			{
				Key:   "vm.max_map_count",
				Value: "2147483647",
			},
		},
	)))

	// E4S/EUS
	treePipeline.AddStage(osbuild.NewDNFConfigStage(osbuild.NewDNFConfigStageOptions(
		[]osbuild.DNFVariable{
			{
				Name:  "releasever",
				Value: t.arch.distro.osVersion,
			},
		},
	)))

	treePipeline = prependKernelCmdlineStage(treePipeline, t, &partitionTable)
	treePipeline.AddStage(osbuild.NewFSTabStage(partitionTable.FSTabStageOptionsV2()))
	kernelVer := kernelVerStr(packageSetSpecs[blueprintPkgsKey], customizations.GetKernel().Name, t.Arch().Name())
	treePipeline.AddStage(bootloaderConfigStage(t, partitionTable, customizations.GetKernel(), kernelVer, false, false))
	// The last stage must be the SELinux stage
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	imagePipeline := liveImagePipeline(treePipeline.Name, diskfile, &partitionTable, t.arch, kernelVer)
	pipelines = append(pipelines, *imagePipeline)
	return pipelines, nil
}

// ec2Pipelines returns pipelines which produce uncompressed EC2 images which are expected to use RHSM for content
func ec2Pipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	return ec2CommonPipelines(t, customizations, options, repos, packageSetSpecs, rng, false, t.Filename())
}

// rhelEc2Pipelines returns pipelines which produce XZ-compressed EC2 images which are expected to use RHUI for content
func rhelEc2Pipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	rawImageFilename := "image.raw"

	pipelines, err := ec2CommonPipelines(t, customizations, options, repos, packageSetSpecs, rng, true, rawImageFilename)
	if err != nil {
		return nil, err
	}

	lastPipeline := pipelines[len(pipelines)-1]
	pipelines = append(pipelines, *xzArchivePipeline(lastPipeline.Name, rawImageFilename, t.Filename()))

	return pipelines, nil
}

// rhelEc2SapPipelines returns pipelines which produce XZ-compressed EC2 SAP images which are expected to use RHUI for content
func rhelEc2SapPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	rawImageFilename := "image.raw"

	pipelines, err := ec2SapPipelines(t, customizations, options, repos, packageSetSpecs, rng, true, rawImageFilename)
	if err != nil {
		return nil, err
	}

	lastPipeline := pipelines[len(pipelines)-1]
	pipelines = append(pipelines, *xzArchivePipeline(lastPipeline.Name, rawImageFilename, t.Filename()))

	return pipelines, nil
}

func tarPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	treePipeline, err := osPipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)
	tarPipeline := osbuild.Pipeline{
		Name:  "root-tar",
		Build: "name:build",
	}
	tarPipeline.AddStage(tarStage("os", "root.tar.xz"))
	pipelines = append(pipelines, tarPipeline)
	return pipelines, nil
}

//makeISORootPath return a path that can be used to address files and folders in
//the root of the iso
func makeISORootPath(p string) string {
	fullpath := path.Join("/run/install/repo", p)
	return fmt.Sprintf("file://%s", fullpath)
}

func edgeInstallerPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))
	installerPackages := packageSetSpecs[installerPkgsKey]
	d := t.arch.distro
	archName := t.arch.name
	kernelVer := kernelVerStr(installerPackages, "kernel", archName)
	ostreeRepoPath := "/ostree/repo"
	payloadStages := ostreePayloadStages(options, ostreeRepoPath)
	kickstartOptions := ostreeKickstartStageOptions(makeISORootPath(ostreeRepoPath), options.OSTree.Ref)
	pipelines = append(pipelines, *anacondaTreePipeline(repos, installerPackages, kernelVer, archName, d.product, d.osVersion, "edge"))
	isolabel := fmt.Sprintf(d.isolabelTmpl, archName)
	pipelines = append(pipelines, *bootISOTreePipeline(kernelVer, archName, d.vendor, d.product, d.osVersion, isolabel, kickstartOptions, payloadStages))
	pipelines = append(pipelines, *bootISOPipeline(t.Filename(), d.isolabelTmpl, archName, false))
	return pipelines, nil
}

func tarInstallerPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	treePipeline, err := osPipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}
	treePipeline.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	pipelines = append(pipelines, *treePipeline)

	kernelPkg := new(rpmmd.PackageSpec)
	installerPackages := packageSetSpecs[installerPkgsKey]
	for _, pkg := range installerPackages {
		if pkg.Name == "kernel" {
			kernelPkg = &pkg
			break
		}
	}
	if kernelPkg == nil {
		return nil, fmt.Errorf("kernel package not found in installer package set")
	}
	kernelVer := fmt.Sprintf("%s-%s.%s", kernelPkg.Version, kernelPkg.Release, kernelPkg.Arch)

	tarPath := "/liveimg.tar"
	tarPayloadStages := []*osbuild.Stage{tarStage("os", tarPath)}
	kickstartOptions := tarKickstartStageOptions(makeISORootPath(tarPath))
	archName := t.arch.name
	d := t.arch.distro
	pipelines = append(pipelines, *anacondaTreePipeline(repos, installerPackages, kernelVer, archName, d.product, d.osVersion, "BaseOS"))
	isolabel := fmt.Sprintf(d.isolabelTmpl, archName)
	pipelines = append(pipelines, *bootISOTreePipeline(kernelVer, archName, d.vendor, d.product, d.osVersion, isolabel, kickstartOptions, tarPayloadStages))
	pipelines = append(pipelines, *bootISOPipeline(t.Filename(), d.isolabelTmpl, t.Arch().Name(), true))
	return pipelines, nil
}

func edgeCorePipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	treePipeline, err := ostreeTreePipeline(repos, packageSetSpecs[osPkgsKey], packageSetSpecs[blueprintPkgsKey], customizations, options, t.enabledServices, t.disabledServices, t.defaultTarget)
	if err != nil {
		return nil, err
	}

	pipelines = append(pipelines, *treePipeline)
	pipelines = append(pipelines, *ostreeCommitPipeline(options, t.arch.distro.osVersion))

	return pipelines, nil
}

func edgeCommitPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines, err := edgeCorePipelines(t, customizations, options, repos, packageSetSpecs)
	if err != nil {
		return nil, err
	}
	tarPipeline := osbuild.Pipeline{
		Name:  "commit-archive",
		Build: "name:build",
	}
	tarPipeline.AddStage(tarStage("ostree-commit", t.Filename()))
	pipelines = append(pipelines, tarPipeline)
	return pipelines, nil
}

func edgeContainerPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines, err := edgeCorePipelines(t, customizations, options, repos, packageSetSpecs)
	if err != nil {
		return nil, err
	}

	nginxConfigPath := "/etc/nginx.conf"
	httpPort := "8080"
	pipelines = append(pipelines, *containerTreePipeline(repos, packageSetSpecs[containerPkgsKey], options, customizations, nginxConfigPath, httpPort))
	pipelines = append(pipelines, *containerPipeline(t, nginxConfigPath, httpPort))
	return pipelines, nil
}

func edgeImagePipelines(t *imageType, filename string, options distro.ImageOptions, rng *rand.Rand) ([]osbuild.Pipeline, string, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	ostreeRepoPath := "/ostree/repo"
	imgName := "image.raw"

	partitionTable, err := t.getPartitionTable(nil, options, rng)
	if err != nil {
		return nil, "", err
	}

	// prepare ostree deployment tree
	treePipeline := ostreeDeployPipeline(t, &partitionTable, ostreeRepoPath, nil, "", rng, options)
	pipelines = append(pipelines, *treePipeline)

	// make raw image from tree
	imagePipeline := liveImagePipeline(treePipeline.Name, imgName, &partitionTable, t.arch, "")
	pipelines = append(pipelines, *imagePipeline)

	// compress image
	xzPipeline := xzArchivePipeline(imagePipeline.Name, imgName, filename)
	pipelines = append(pipelines, *xzPipeline)

	return pipelines, xzPipeline.Name, nil
}

func edgeRawImagePipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))

	imgName := t.filename

	// create the raw image
	imagePipelines, _, err := edgeImagePipelines(t, imgName, options, rng)
	if err != nil {
		return nil, err
	}

	pipelines = append(pipelines, imagePipelines...)

	return pipelines, nil
}

func buildPipeline(repos []rpmmd.RepoConfig, buildPackageSpecs []rpmmd.PackageSpec, runner string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "build"
	p.Runner = runner
	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(buildPackageSpecs)))
	p.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(true)))
	return p
}

func osPipeline(repos []rpmmd.RepoConfig, packages []rpmmd.PackageSpec, bpPackages []rpmmd.PackageSpec, c *blueprint.Customizations, options distro.ImageOptions, enabledServices, disabledServices []string, defaultTarget string) (*osbuild.Pipeline, error) {
	p := new(osbuild.Pipeline)
	p.Name = "os"
	p.Build = "name:build"
	packages = append(packages, bpPackages...)
	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(packages)))
	p.AddStage(osbuild.NewFixBLSStage(&osbuild.FixBLSStageOptions{}))
	language, keyboard := c.GetPrimaryLocale()
	if language != nil {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: *language}))
	} else {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: "en_US.UTF-8"}))
	}
	if keyboard != nil {
		p.AddStage(osbuild.NewKeymapStage(&osbuild.KeymapStageOptions{Keymap: *keyboard}))
	}
	if hostname := c.GetHostname(); hostname != nil {
		p.AddStage(osbuild.NewHostnameStage(&osbuild.HostnameStageOptions{Hostname: *hostname}))
	}

	timezone, ntpServers := c.GetTimezoneSettings()
	if timezone != nil {
		p.AddStage(osbuild.NewTimezoneStage(&osbuild.TimezoneStageOptions{Zone: *timezone}))
	} else {
		p.AddStage(osbuild.NewTimezoneStage(&osbuild.TimezoneStageOptions{Zone: "America/New_York"}))
	}

	if len(ntpServers) > 0 {
		p.AddStage(osbuild.NewChronyStage(&osbuild.ChronyStageOptions{Timeservers: ntpServers}))
	}

	if groups := c.GetGroups(); len(groups) > 0 {
		p.AddStage(osbuild.NewGroupsStage(groupStageOptions(groups)))
	}

	if users := c.GetUsers(); len(users) > 0 {
		userOptions, err := userStageOptions(users)
		if err != nil {
			return nil, err
		}
		p.AddStage(osbuild.NewUsersStage(userOptions))
	}

	if services := c.GetServices(); services != nil || enabledServices != nil || disabledServices != nil || defaultTarget != "" {
		p.AddStage(osbuild.NewSystemdStage(systemdStageOptions(enabledServices, disabledServices, services, defaultTarget)))
	}

	if firewall := c.GetFirewall(); firewall != nil {
		p.AddStage(osbuild.NewFirewallStage(firewallStageOptions(firewall)))
	}

	// These are the current defaults for the sysconfig stage. This can be changed to be image type exclusive if different configs are needed.
	p.AddStage(osbuild.NewSysconfigStage(&osbuild.SysconfigStageOptions{
		Kernel: osbuild.SysconfigKernelOptions{
			UpdateDefault: true,
			DefaultKernel: "kernel",
		},
		Network: osbuild.SysconfigNetworkOptions{
			Networking: true,
			NoZeroConf: true,
		},
	}))

	if options.Subscription != nil {
		commands := []string{
			fmt.Sprintf("/usr/sbin/subscription-manager register --org=%s --activationkey=%s --serverurl %s --baseurl %s", options.Subscription.Organization, options.Subscription.ActivationKey, options.Subscription.ServerUrl, options.Subscription.BaseUrl),
		}
		if options.Subscription.Insights {
			commands = append(commands, "/usr/bin/insights-client --register")
		}

		p.AddStage(osbuild.NewFirstBootStage(&osbuild.FirstBootStageOptions{
			Commands:       commands,
			WaitForNetwork: true,
		},
		))
	}
	return p, nil
}

func ostreeTreePipeline(repos []rpmmd.RepoConfig, packages []rpmmd.PackageSpec, bpPackages []rpmmd.PackageSpec, c *blueprint.Customizations, options distro.ImageOptions, enabledServices, disabledServices []string, defaultTarget string) (*osbuild.Pipeline, error) {
	p := new(osbuild.Pipeline)
	p.Name = "ostree-tree"
	p.Build = "name:build"

	packages = append(packages, bpPackages...)

	if options.OSTree.Parent != "" && options.OSTree.URL != "" {
		p.AddStage(osbuild.NewOSTreePasswdStage("org.osbuild.source", options.OSTree.Parent))
	}

	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(packages)))
	p.AddStage(osbuild.NewFixBLSStage(&osbuild.FixBLSStageOptions{}))
	language, keyboard := c.GetPrimaryLocale()
	if language != nil {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: *language}))
	} else {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: "en_US.UTF-8"}))
	}
	if keyboard != nil {
		p.AddStage(osbuild.NewKeymapStage(&osbuild.KeymapStageOptions{Keymap: *keyboard}))
	}
	if hostname := c.GetHostname(); hostname != nil {
		p.AddStage(osbuild.NewHostnameStage(&osbuild.HostnameStageOptions{Hostname: *hostname}))
	}

	timezone, ntpServers := c.GetTimezoneSettings()
	if timezone != nil {
		p.AddStage(osbuild.NewTimezoneStage(&osbuild.TimezoneStageOptions{Zone: *timezone}))
	} else {
		p.AddStage(osbuild.NewTimezoneStage(&osbuild.TimezoneStageOptions{Zone: "America/New_York"}))
	}

	if len(ntpServers) > 0 {
		p.AddStage(osbuild.NewChronyStage(&osbuild.ChronyStageOptions{Timeservers: ntpServers}))
	}

	if groups := c.GetGroups(); len(groups) > 0 {
		p.AddStage(osbuild.NewGroupsStage(groupStageOptions(groups)))
	}

	if users := c.GetUsers(); len(users) > 0 {
		userOptions, err := userStageOptions(users)
		if err != nil {
			return nil, err
		}
		p.AddStage(osbuild.NewUsersStage(userOptions))
		p.AddStage(osbuild.NewFirstBootStage(usersFirstBootOptions(userOptions)))
	}

	if services := c.GetServices(); services != nil || enabledServices != nil || disabledServices != nil || defaultTarget != "" {
		p.AddStage(osbuild.NewSystemdStage(systemdStageOptions(enabledServices, disabledServices, services, defaultTarget)))
	}

	if firewall := c.GetFirewall(); firewall != nil {
		p.AddStage(osbuild.NewFirewallStage(firewallStageOptions(firewall)))
	}

	// These are the current defaults for the sysconfig stage. This can be changed to be image type exclusive if different configs are needed.
	p.AddStage(osbuild.NewSysconfigStage(&osbuild.SysconfigStageOptions{
		Kernel: osbuild.SysconfigKernelOptions{
			UpdateDefault: true,
			DefaultKernel: "kernel",
		},
		Network: osbuild.SysconfigNetworkOptions{
			Networking: true,
			NoZeroConf: true,
		},
	}))

	if options.Subscription != nil {
		commands := []string{
			fmt.Sprintf("/usr/sbin/subscription-manager register --org=%s --activationkey=%s --serverurl %s --baseurl %s", options.Subscription.Organization, options.Subscription.ActivationKey, options.Subscription.ServerUrl, options.Subscription.BaseUrl),
		}
		if options.Subscription.Insights {
			commands = append(commands, "/usr/bin/insights-client --register")
		}

		p.AddStage(osbuild.NewFirstBootStage(&osbuild.FirstBootStageOptions{
			Commands:       commands,
			WaitForNetwork: true,
		},
		))
	}

	p.AddStage(osbuild.NewSELinuxStage(selinuxStageOptions(false)))
	p.AddStage(osbuild.NewOSTreePrepTreeStage(&osbuild.OSTreePrepTreeStageOptions{
		EtcGroupMembers: []string{
			// NOTE: We may want to make this configurable.
			"wheel", "docker",
		},
	}))
	return p, nil
}
func ostreeCommitPipeline(options distro.ImageOptions, osVersion string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "ostree-commit"
	p.Build = "name:build"
	p.AddStage(osbuild.NewOSTreeInitStage(&osbuild.OSTreeInitStageOptions{Path: "/repo"}))

	commitStageInput := new(osbuild.OSTreeCommitStageInput)
	commitStageInput.Type = "org.osbuild.tree"
	commitStageInput.Origin = "org.osbuild.pipeline"
	commitStageInput.References = osbuild.OSTreeCommitStageReferences{"name:ostree-tree"}

	p.AddStage(osbuild.NewOSTreeCommitStage(
		&osbuild.OSTreeCommitStageOptions{
			Ref:       options.OSTree.Ref,
			OSVersion: osVersion,
			Parent:    options.OSTree.Parent,
		},
		&osbuild.OSTreeCommitStageInputs{Tree: commitStageInput}),
	)
	return p
}

func tarStage(source, filename string) *osbuild.Stage {
	tree := new(osbuild.TarStageInput)
	tree.Type = "org.osbuild.tree"
	tree.Origin = "org.osbuild.pipeline"
	tree.References = []string{"name:" + source}
	return osbuild.NewTarStage(&osbuild.TarStageOptions{Filename: filename}, &osbuild.TarStageInputs{Tree: tree})
}

func containerTreePipeline(repos []rpmmd.RepoConfig, packages []rpmmd.PackageSpec, options distro.ImageOptions, c *blueprint.Customizations, nginxConfigPath, listenPort string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "container-tree"
	p.Build = "name:build"
	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(packages)))
	language, _ := c.GetPrimaryLocale()
	if language != nil {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: *language}))
	} else {
		p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: "en_US"}))
	}

	htmlRoot := "/usr/share/nginx/html"
	repoPath := filepath.Join(htmlRoot, "repo")
	p.AddStage(osbuild.NewOSTreeInitStage(&osbuild.OSTreeInitStageOptions{Path: repoPath}))

	p.AddStage(osbuild.NewOSTreePullStage(
		&osbuild.OSTreePullStageOptions{Repo: repoPath},
		ostreePullStageInputs("org.osbuild.pipeline", "name:ostree-commit", options.OSTree.Ref),
	))

	// make nginx log directory world writeable, otherwise nginx can't start in
	// an unprivileged container
	p.AddStage(osbuild.NewChmodStage(chmodStageOptions("/var/log/nginx", "o+w", true)))

	p.AddStage(osbuild.NewNginxConfigStage(nginxConfigStageOptions(nginxConfigPath, htmlRoot, listenPort)))
	return p
}

func containerPipeline(t *imageType, nginxConfigPath, listenPort string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "container"
	p.Build = "name:build"
	options := &osbuild.OCIArchiveStageOptions{
		Architecture: t.arch.Name(),
		Filename:     t.Filename(),
		Config: &osbuild.OCIArchiveConfig{
			Cmd:          []string{"nginx", "-c", nginxConfigPath},
			ExposedPorts: []string{listenPort},
		},
	}
	baseInput := new(osbuild.OCIArchiveStageInput)
	baseInput.Type = "org.osbuild.tree"
	baseInput.Origin = "org.osbuild.pipeline"
	baseInput.References = []string{"name:container-tree"}
	inputs := &osbuild.OCIArchiveStageInputs{Base: baseInput}
	p.AddStage(osbuild.NewOCIArchiveStage(options, inputs))
	return p
}

func ostreePayloadStages(options distro.ImageOptions, ostreeRepoPath string) []*osbuild.Stage {
	stages := make([]*osbuild.Stage, 0)

	// ostree commit payload
	stages = append(stages, osbuild.NewOSTreeInitStage(&osbuild.OSTreeInitStageOptions{Path: ostreeRepoPath}))
	stages = append(stages, osbuild.NewOSTreePullStage(
		&osbuild.OSTreePullStageOptions{Repo: ostreeRepoPath},
		ostreePullStageInputs("org.osbuild.source", options.OSTree.Parent, options.OSTree.Ref),
	))

	return stages
}

func edgeSimplifiedInstallerPipelines(t *imageType, customizations *blueprint.Customizations, options distro.ImageOptions, repos []rpmmd.RepoConfig, packageSetSpecs map[string][]rpmmd.PackageSpec, rng *rand.Rand) ([]osbuild.Pipeline, error) {
	pipelines := make([]osbuild.Pipeline, 0)
	pipelines = append(pipelines, *buildPipeline(repos, packageSetSpecs[buildPkgsKey], t.arch.distro.runner))
	installerPackages := packageSetSpecs[installerPkgsKey]
	kernelVer := kernelVerStr(installerPackages, "kernel", t.Arch().Name())
	imgName := "disk.img.xz"
	installDevice := customizations.GetInstallationDevice()

	// create the raw image
	imagePipelines, imgPipelineName, err := edgeImagePipelines(t, imgName, options, rng)
	if err != nil {
		return nil, err
	}

	pipelines = append(pipelines, imagePipelines...)

	// create boot ISO with raw image
	d := t.arch.distro
	archName := t.arch.name
	installerTreePipeline := simplifiedInstallerTreePipeline(repos, installerPackages, kernelVer, archName, d.product, d.osVersion, "edge")
	isolabel := fmt.Sprintf(d.isolabelTmpl, archName)
	efibootTreePipeline := simplifiedInstallerEFIBootTreePipeline(installDevice, kernelVer, archName, d.vendor, d.product, d.osVersion, isolabel)
	bootISOTreePipeline := simplifiedInstallerBootISOTreePipeline(imgPipelineName, kernelVer)

	pipelines = append(pipelines, *installerTreePipeline, *efibootTreePipeline, *bootISOTreePipeline)
	pipelines = append(pipelines, *bootISOPipeline(t.Filename(), d.isolabelTmpl, t.Arch().Name(), false))

	return pipelines, nil
}

func simplifiedInstallerBootISOTreePipeline(archivePipelineName, kver string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "bootiso-tree"
	p.Build = "name:build"

	p.AddStage(osbuild.NewCopyStageSimple(
		&osbuild.CopyStageOptions{
			Paths: []osbuild.CopyStagePath{
				{
					From: "input://file/disk.img.xz",
					To:   "tree:///disk.img.xz",
				},
			},
		},
		osbuild.NewFilesInputs(osbuild.NewFilesInputReferencesPipeline(archivePipelineName, "disk.img.xz")),
	))

	p.AddStage(osbuild.NewMkdirStage(
		&osbuild.MkdirStageOptions{
			Paths: []osbuild.Path{
				{
					Path: "images",
				},
				{
					Path: "images/pxeboot",
				},
			},
		},
	))

	var sectorSize uint64 = 512
	pt := disk.PartitionTable{
		Size: 20971520,
		Partitions: []disk.Partition{
			{
				Start: 0,
				Size:  20971520 / sectorSize,
				Filesystem: &disk.Filesystem{
					Type:       "vfat",
					Mountpoint: "/",
				},
			},
		},
	}

	filename := "images/efiboot.img"
	loopback := osbuild.NewLoopbackDevice(&osbuild.LoopbackDeviceOptions{Filename: filename})
	p.AddStage(osbuild.NewTruncateStage(&osbuild.TruncateStageOptions{Filename: filename, Size: fmt.Sprintf("%d", pt.Size)}))

	for _, stage := range mkfsStages(&pt, loopback) {
		p.AddStage(stage)
	}

	inputName := "root-tree"
	copyInputs := copyPipelineTreeInputs(inputName, "efiboot-tree")
	copyOptions, copyDevices, copyMounts := copyFSTreeOptions(inputName, "efiboot-tree", &pt, loopback)
	p.AddStage(osbuild.NewCopyStage(copyOptions, copyInputs, copyDevices, copyMounts))

	inputName = "coi"
	copyInputs = copyPipelineTreeInputs(inputName, "coi-tree")
	p.AddStage(osbuild.NewCopyStageSimple(
		&osbuild.CopyStageOptions{
			Paths: []osbuild.CopyStagePath{
				{
					From: fmt.Sprintf("input://%s/boot/vmlinuz-%s", inputName, kver),
					To:   "tree:///images/pxeboot/vmlinuz",
				},
				{
					From: fmt.Sprintf("input://%s/boot/initramfs-%s.img", inputName, kver),
					To:   "tree:///images/pxeboot/initrd.img",
				},
			},
		},
		copyInputs,
	))

	inputName = "efi-tree"
	copyInputs = copyPipelineTreeInputs(inputName, "efiboot-tree")
	p.AddStage(osbuild.NewCopyStageSimple(
		&osbuild.CopyStageOptions{
			Paths: []osbuild.CopyStagePath{
				{
					From: fmt.Sprintf("input://%s/EFI", inputName),
					To:   "tree:///",
				},
			},
		},
		copyInputs,
	))

	return p
}

func simplifiedInstallerEFIBootTreePipeline(installDevice, kernelVer, arch, vendor, product, osVersion, isolabel string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "efiboot-tree"
	p.Build = "name:build"
	p.AddStage(osbuild.NewGrubISOStage(grubISOStageOptions(installDevice, kernelVer, arch, vendor, product, osVersion, isolabel)))
	return p
}

func simplifiedInstallerTreePipeline(repos []rpmmd.RepoConfig, packages []rpmmd.PackageSpec, kernelVer, arch, product, osVersion, variant string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "coi-tree"
	p.Build = "name:build"
	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(packages)))
	p.AddStage(osbuild.NewBuildstampStage(buildStampStageOptions(arch, product, osVersion, variant)))
	p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: "en_US.UTF-8"}))
	p.AddStage(osbuild.NewSystemdStage(systemdStageOptions([]string{"coreos-installer"}, nil, nil, "")))
	p.AddStage(osbuild.NewDracutStage(dracutStageOptions(kernelVer, arch, []string{"rdcore"})))

	return p
}

func ostreeDeployPipeline(
	t *imageType,
	pt *disk.PartitionTable,
	repoPath string,
	kernel *blueprint.KernelCustomization,
	kernelVer string,
	rng *rand.Rand,
	options distro.ImageOptions,
) *osbuild.Pipeline {

	p := new(osbuild.Pipeline)
	p.Name = "image-tree"
	p.Build = "name:build"
	osname := "redhat"

	p.AddStage(osbuild.OSTreeInitFsStage())
	p.AddStage(osbuild.NewOSTreePullStage(
		&osbuild.OSTreePullStageOptions{Repo: repoPath},
		ostreePullStageInputs("org.osbuild.source", options.OSTree.Parent, options.OSTree.Ref),
	))
	p.AddStage(osbuild.NewOSTreeOsInitStage(
		&osbuild.OSTreeOsInitStageOptions{
			OSName: osname,
		},
	))
	p.AddStage(osbuild.NewOSTreeConfigStage(ostreeConfigStageOptions(repoPath, true)))
	p.AddStage(osbuild.NewMkdirStage(efiMkdirStageOptions()))
	p.AddStage(osbuild.NewOSTreeDeployStage(
		&osbuild.OSTreeDeployStageOptions{
			OsName: osname,
			Ref:    options.OSTree.Ref,
			Mounts: []string{"/boot", "/boot/efi"},
			Rootfs: osbuild.Rootfs{
				Label: "root",
			},
			KernelOpts: []string{
				"console=tty0",
				"console=ttyS0",
			},
		},
	))
	p.AddStage(osbuild.NewOSTreeFillvarStage(
		&osbuild.OSTreeFillvarStageOptions{
			Deployment: osbuild.OSTreeDeployment{
				OSName: osname,
				Ref:    options.OSTree.Ref,
			},
		},
	))

	fstabOptions := pt.FSTabStageOptionsV2()
	fstabOptions.OSTree = &osbuild.OSTreeFstab{
		Deployment: osbuild.OSTreeDeployment{
			OSName: osname,
			Ref:    options.OSTree.Ref,
		},
	}
	p.AddStage(osbuild.NewFSTabStage(fstabOptions))

	// TODO: Add users?

	p.AddStage(bootloaderConfigStage(t, *pt, kernel, kernelVer, true, true))

	p.AddStage(osbuild.NewOSTreeSelinuxStage(
		&osbuild.OSTreeSelinuxStageOptions{
			Deployment: osbuild.OSTreeDeployment{
				OSName: osname,
				Ref:    options.OSTree.Ref,
			},
		},
	))
	return p
}

func anacondaTreePipeline(repos []rpmmd.RepoConfig, packages []rpmmd.PackageSpec, kernelVer, arch, product, osVersion, variant string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "anaconda-tree"
	p.Build = "name:build"
	p.AddStage(osbuild.NewRPMStage(rpmStageOptions(repos), rpmStageInputs(packages)))
	p.AddStage(osbuild.NewBuildstampStage(buildStampStageOptions(arch, product, osVersion, variant)))
	p.AddStage(osbuild.NewLocaleStage(&osbuild.LocaleStageOptions{Language: "en_US.UTF-8"}))

	rootPassword := ""
	rootUser := osbuild.UsersStageOptionsUser{
		Password: &rootPassword,
	}

	installUID := 0
	installGID := 0
	installHome := "/root"
	installShell := "/usr/libexec/anaconda/run-anaconda"
	installPassword := ""
	installUser := osbuild.UsersStageOptionsUser{
		UID:      &installUID,
		GID:      &installGID,
		Home:     &installHome,
		Shell:    &installShell,
		Password: &installPassword,
	}
	usersStageOptions := &osbuild.UsersStageOptions{
		Users: map[string]osbuild.UsersStageOptionsUser{
			"root":    rootUser,
			"install": installUser,
		},
	}

	p.AddStage(osbuild.NewUsersStage(usersStageOptions))
	p.AddStage(osbuild.NewAnacondaStage(anacondaStageOptions()))
	p.AddStage(osbuild.NewLoraxScriptStage(loraxScriptStageOptions(arch)))
	p.AddStage(osbuild.NewDracutStage(dracutStageOptions(kernelVer, arch, []string{
		"anaconda",
	})))

	return p
}

func bootISOTreePipeline(kernelVer, arch, vendor, product, osVersion, isolabel string, ksOptions *osbuild.KickstartStageOptions, payloadStages []*osbuild.Stage) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "bootiso-tree"
	p.Build = "name:build"

	p.AddStage(osbuild.NewBootISOMonoStage(bootISOMonoStageOptions(kernelVer, arch, vendor, product, osVersion, isolabel), bootISOMonoStageInputs()))
	p.AddStage(osbuild.NewKickstartStage(ksOptions))
	p.AddStage(osbuild.NewDiscinfoStage(discinfoStageOptions(arch)))

	for _, stage := range payloadStages {
		p.AddStage(stage)
	}

	return p
}
func bootISOPipeline(filename, isolabel, arch string, isolinux bool) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "bootiso"
	p.Build = "name:build"

	p.AddStage(osbuild.NewXorrisofsStage(xorrisofsStageOptions(filename, isolabel, arch, isolinux), xorrisofsStageInputs("bootiso-tree")))
	p.AddStage(osbuild.NewImplantisomd5Stage(&osbuild.Implantisomd5StageOptions{Filename: filename}))

	return p
}

func liveImagePipeline(inputPipelineName string, outputFilename string, pt *disk.PartitionTable, arch *architecture, kernelVer string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "image"
	p.Build = "name:build"

	p.AddStage(osbuild.NewTruncateStage(&osbuild.TruncateStageOptions{Filename: outputFilename, Size: fmt.Sprintf("%d", pt.Size)}))
	sfOptions := sfdiskStageOptions(pt)
	loopback := osbuild.NewLoopbackDevice(&osbuild.LoopbackDeviceOptions{Filename: outputFilename})
	p.AddStage(osbuild.NewSfdiskStage(sfOptions, loopback))

	for _, stage := range mkfsStages(pt, loopback) {
		p.AddStage(stage)
	}

	inputName := "root-tree"
	copyOptions, copyDevices, copyMounts := copyFSTreeOptions(inputName, inputPipelineName, pt, loopback)
	copyInputs := copyPipelineTreeInputs(inputName, inputPipelineName)
	p.AddStage(osbuild.NewCopyStage(copyOptions, copyInputs, copyDevices, copyMounts))
	p.AddStage(bootloaderInstStage(outputFilename, pt, arch, kernelVer, copyDevices, copyMounts, loopback))
	return p
}

func xzArchivePipeline(inputPipelineName, inputFilename, outputFilename string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = "archive"
	p.Build = "name:build"

	p.AddStage(osbuild.NewXzStage(
		osbuild.NewXzStageOptions(outputFilename),
		osbuild.NewFilesInputs(osbuild.NewFilesInputReferencesPipeline(inputPipelineName, inputFilename)),
	))

	return p
}

// mkfsStages generates a list of org.osbuild.mkfs.* stages based on a
// partition table description for a single device node
func mkfsStages(pt *disk.PartitionTable, device *osbuild.Device) []*osbuild.Stage {
	stages := make([]*osbuild.Stage, 0, len(pt.Partitions))

	// assume loopback device for simplicity since it's the only one currently supported
	// panic if the conversion fails
	devOptions, ok := device.Options.(*osbuild.LoopbackDeviceOptions)
	if !ok {
		panic("mkfsStages: failed to convert device options to loopback options")
	}

	for _, p := range pt.Partitions {
		if p.Filesystem == nil {
			// no filesystem for partition (e.g., BIOS boot)
			continue
		}
		var stage *osbuild.Stage
		stageDevice := osbuild.NewLoopbackDevice(
			&osbuild.LoopbackDeviceOptions{
				Filename: devOptions.Filename,
				Start:    p.Start,
				Size:     p.Size,
			},
		)
		switch p.Filesystem.Type {
		case "xfs":
			options := &osbuild.MkfsXfsStageOptions{
				UUID:  p.Filesystem.UUID,
				Label: p.Filesystem.Label,
			}
			stage = osbuild.NewMkfsXfsStage(options, stageDevice)
		case "vfat":
			options := &osbuild.MkfsFATStageOptions{
				VolID: strings.Replace(p.Filesystem.UUID, "-", "", -1),
			}
			stage = osbuild.NewMkfsFATStage(options, stageDevice)
		case "btrfs":
			options := &osbuild.MkfsBtrfsStageOptions{
				UUID:  p.Filesystem.UUID,
				Label: p.Filesystem.Label,
			}
			stage = osbuild.NewMkfsBtrfsStage(options, stageDevice)
		case "ext4":
			options := &osbuild.MkfsExt4StageOptions{
				UUID:  p.Filesystem.UUID,
				Label: p.Filesystem.Label,
			}
			stage = osbuild.NewMkfsExt4Stage(options, stageDevice)
		default:
			panic("unknown fs type " + p.Type)
		}
		stages = append(stages, stage)
	}
	return stages
}

func qemuPipeline(inputPipelineName, inputFilename, outputFilename, format, qcow2Compat string) *osbuild.Pipeline {
	p := new(osbuild.Pipeline)
	p.Name = format
	p.Build = "name:build"

	qemuStage := osbuild.NewQEMUStage(qemuStageOptions(outputFilename, format, qcow2Compat), qemuStageInputs(inputPipelineName, inputFilename))
	p.AddStage(qemuStage)
	return p
}

func bootloaderConfigStage(t *imageType, partitionTable disk.PartitionTable, kernel *blueprint.KernelCustomization, kernelVer string, install, greenboot bool) *osbuild.Stage {
	if t.arch.name == distro.S390xArchName {
		return osbuild.NewZiplStage(new(osbuild.ZiplStageOptions))
	}

	kernelOptions := t.kernelOptions
	uefi := t.supportsUEFI()
	legacy := t.arch.legacy

	options := grub2StageOptions(partitionTable.RootPartition(), partitionTable.BootPartition(), kernelOptions, kernel, kernelVer, uefi, legacy, t.arch.distro.vendor, install)
	options.Greenboot = greenboot

	return osbuild.NewGRUB2Stage(options)
}

func bootloaderInstStage(filename string, pt *disk.PartitionTable, arch *architecture, kernelVer string, devices *osbuild.Devices, mounts *osbuild.Mounts, disk *osbuild.Device) *osbuild.Stage {
	platform := arch.legacy
	if platform != "" {
		return osbuild.NewGrub2InstStage(grub2InstStageOptions(filename, pt, platform))
	}

	if arch.name == distro.S390xArchName {
		return osbuild.NewZiplInstStage(ziplInstStageOptions(kernelVer, pt), disk, devices, mounts)
	}

	return nil
}

func kernelVerStr(pkgs []rpmmd.PackageSpec, kernelName, arch string) string {
	kernelPkg := new(rpmmd.PackageSpec)
	for _, pkg := range pkgs {
		if pkg.Name == kernelName {
			kernelPkg = &pkg
			break
		}
	}
	if kernelPkg == nil {
		panic(fmt.Sprintf("kernel package %q not found", kernelName))
	}
	return fmt.Sprintf("%s-%s.%s", kernelPkg.Version, kernelPkg.Release, kernelPkg.Arch)
}