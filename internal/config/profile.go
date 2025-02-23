package config

import "github.com/spf13/viper"

// Profile is a struct that holds the profile information
type Profile struct {
	Name     string
	Server   string
	Username string
	Password string
}

func GetProfiles() []Profile {
	// get profiles from config
	var profiles []Profile
	viper.UnmarshalKey("profiles", &profiles)

	return profiles
}

func GetDefaultProfile() Profile {
	// get profiles from config
	var profiles []Profile
	viper.UnmarshalKey("profiles", &profiles)

	defaultProfile := viper.GetString("default_profile")

	for _, p := range profiles {
		if p.Name == defaultProfile {
			return p
		}
	}

	return Profile{}
}

func GetDefaultProfileName() string {
	return viper.GetString("default_profile")
}

func SetDefaultProfile(name string, commit bool) error {
	// get profiles from config
	var profiles []Profile
	viper.UnmarshalKey("profiles", &profiles)

	// make sure profile exists
	for _, p := range profiles {
		if p.Name == name {
			viper.Set("default_profile", name)
			if commit {
				viper.WriteConfig()
			}
			return nil
		}
	}
	return &ProfileNotFoundError{name}
}

func AddProfile(name string, isDefault bool, server string, username string, password string) *Profile {
	// get profiles from config
	var profiles []Profile
	viper.UnmarshalKey("profiles", &profiles)

	// add new profile
	profiles = append(profiles, Profile{
		Name:     name,
		Server:   server,
		Username: username,
		Password: password,
	})

	// if default, set all other profiles to non default
	if isDefault {
		viper.Set("default_profile", name)
	}

	viper.Set("profiles", profiles)
	viper.WriteConfig()

	return &profiles[len(profiles)-1]
}

func RemoveProfile(name string) {
	// get profiles from config
	var profiles []Profile
	viper.UnmarshalKey("profiles", &profiles)

	// remove profile
	for i, p := range profiles {
		if p.Name == name {
			profiles = append(profiles[:i], profiles[i+1:]...)
		}
	}

	viper.Set("profiles", profiles)
	viper.WriteConfig()
}

// profile not found error definition
type ProfileNotFoundError struct {
	Name string
}

func (e *ProfileNotFoundError) Error() string {
	return "Profile not found: " + e.Name
}
