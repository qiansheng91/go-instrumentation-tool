package config

type InstrumentationInfo interface {
	Plugins() []InjectPlugin
	WeavePackages() map[string]InjectPlugin
}

type instrumentationInfoImpl struct {
	config *configuration
}

func (i instrumentationInfoImpl) WeavePackages() map[string]InjectPlugin {
	deps := make(map[string]InjectPlugin, 0)

	for _, p := range i.config.Plugins {
		for _, tp := range p.Target_packages {
			deps[tp.PackagePath] = newInjectPluginImpl(p)
		}
	}

	return deps
}

func (i instrumentationInfoImpl) Plugins() []InjectPlugin {
	plugins := make([]InjectPlugin, 0)

	for _, p := range i.config.Plugins {
		plugins = append(plugins, newInjectPluginImpl(p))
	}

	return plugins
}

func newInstrumentationInfo(config *configuration) InstrumentationInfo {
	return &instrumentationInfoImpl{
		config: config,
	}
}
