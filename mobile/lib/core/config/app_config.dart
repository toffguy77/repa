enum Flavor { dev, staging, prod }

class AppConfig {
  final Flavor flavor;
  final String apiBaseUrl;
  final String appUrl;

  const AppConfig({
    required this.flavor,
    required this.apiBaseUrl,
    required this.appUrl,
  });

  static const dev = AppConfig(
    flavor: Flavor.dev,
    apiBaseUrl: 'http://localhost:3000/api/v1',
    appUrl: 'http://localhost:3000',
  );

  static const staging = AppConfig(
    flavor: Flavor.staging,
    apiBaseUrl: 'https://staging-api.repa.app/api/v1',
    appUrl: 'https://staging.repa.app',
  );

  static const prod = AppConfig(
    flavor: Flavor.prod,
    apiBaseUrl: 'https://api.repa.app/api/v1',
    appUrl: 'https://repa.app',
  );

  static AppConfig fromFlavor(String name) {
    switch (name) {
      case 'staging':
        return staging;
      case 'prod':
        return prod;
      default:
        return dev;
    }
  }

  bool get isDev => flavor == Flavor.dev;
  bool get isStaging => flavor == Flavor.staging;
  bool get isProd => flavor == Flavor.prod;
}
