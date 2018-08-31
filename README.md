# go-server
An example Go Application with modular components


The application uses a central component registry which components self register against.

Each component then uses dependencie Injection so that it can simply list the dependencies needed and they will be populated at startup.

Finally, Viper is used for configuration management.  Each component can specify default values, and within their Init() function can parse and validate settings.

