package tunnel

import (
	/* 	"crypto/ecdh"
	   	"crypto/rand" */

	"fmt"
	"wire-garden-server/internal/log"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type TunnelConfig struct {
	tunnelName string
	privateKey *wgtypes.Key
	publicKey  *wgtypes.Key
	listenPort int
}

var tunnels = make(map[string]*netlink.GenericLink)

func Bootstrap() (err error) {
	log.Logger.Info("Creating tunnel")

	tunnelConfig := &TunnelConfig{
		tunnelName: "test",
		listenPort: 51820,
	}

	/* privateKey, publicKey, err := genKeyPair() */
	_, _, err = genKeyPair()
	if err != nil {
		return err
	}

	link, err := createLink(tunnelConfig)
	if err != nil {
		return err
	}

	err = upInterface(link)
	if err != nil {
		return err
	}

	configureInterface(tunnelConfig)

	return nil
}

func Teardown() error {
	isErrorHappened := false

	log.Logger.Info("Tearing down managed tunnels")

	for _, link := range tunnels {
		if err := netlink.LinkSetDown(link); err != nil {
			isErrorHappened = false
			log.Logger.Error(
				fmt.Sprintf("Failed to stop interface %s", link.Name),
				zap.String("error", err.Error()),
			)
		} else {
			if err := netlink.LinkDel(link); err != nil {
				log.Logger.Error(
					fmt.Sprintf("Failed to delete interface %s", link.Name),
					zap.String("error", err.Error()),
				)
			}
		}
	}

	if isErrorHappened {
		return fmt.Errorf("Errors happened during teardown, see logs above for details")
	} else {
		return nil
	}
}

func genKeyPair() (*wgtypes.Key, *wgtypes.Key, error) {
	/* 	log.Logger.Info("Generating keypair")

	   	curve := ecdh.X25519()
	   	key, _ := curve.GenerateKey(rand.Reader)

	   	privateKey, err := curve.NewPrivateKey(key.Bytes())
	   	if err != nil {
	   		log.Logger.Error("Failed to generate private key")
	   		return "", "", err
	   	}

	   	privateKeyEncoded = b64.StdEncoding.EncodeToString(privateKey.Bytes())

	   	publicKey := privateKey.PublicKey()
	   	publicKeyEncoded = b64.StdEncoding.EncodeToString(publicKey.Bytes())

	   	return privateKeyEncoded, publicKeyEncoded, nil */
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Logger.Error("Failed to generate private key")
		return nil, nil, err
	}

	publicKey := privateKey.PublicKey()

	return &privateKey, &publicKey, nil
}

func createLink(config *TunnelConfig) (link *netlink.GenericLink, err error) {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = config.tunnelName

	if _, exists := tunnels[config.tunnelName]; exists {
		return nil, fmt.Errorf("Interface %s already exists", config.tunnelName)
	}

	if oldLink, err := netlink.LinkByName(config.tunnelName); err == nil && oldLink != nil {
		genericLink, ok := oldLink.(*netlink.GenericLink)

		if !ok {
			return nil, fmt.Errorf("Existing link is not of type *netlink.GenericLink")
		}

		if genericLink.Type() != "wireguard" {
			return nil, fmt.Errorf("Existing link is not of type wireguard")
		}

		link = genericLink
	} else {
		link = &netlink.GenericLink{
			LinkAttrs: linkAttrs,
			LinkType:  "wireguard",
		}

		if err := netlink.LinkAdd(link); err != nil {
			// TODO: Check if module exists /sys/module/wireguard

			log.Logger.Error(
				"Failed to create netlink",
				zap.String("error", err.Error()),
			)

			return nil, err
		}
	}

	tunnels[config.tunnelName] = link

	return link, nil
}

func destroyInterface() error {
	return nil
}

func configureInterface(tunnelConfig *TunnelConfig) error {
	wg, err := wgctrl.New()

	if err != nil {
		log.Logger.Error(
			"Failed to create client to configure wireguard interface",
			zap.String("error", err.Error()),
		)
		return err
	}

	defer wg.Close()

	_, err = wg.Device(tunnelConfig.tunnelName)

	if err != nil {
		log.Logger.Error(
			fmt.Sprintf("Interface '%s' not found (%v)", tunnelConfig.tunnelName, err),
		)

		return err
	}

	wgConfig := wgtypes.Config{
		PrivateKey:   tunnelConfig.privateKey,
		ListenPort:   &tunnelConfig.listenPort,
		ReplacePeers: false,
	}

	err = wg.ConfigureDevice(tunnelConfig.tunnelName, wgConfig)
	if err != nil {
		log.Logger.Error(
			fmt.Sprintf("Failed to configure interface '%s' (%v)", tunnelConfig.tunnelName, err),
		)

		return err
	}

	return nil
}

func upInterface(link *netlink.GenericLink) error {
	return nil
}

func downInterface() error {
	return nil
}
