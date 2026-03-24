import '../config/app_config.dart';

class Env {
  static const _flavor = String.fromEnvironment('FLAVOR', defaultValue: 'dev');
  static final config = AppConfig.fromFlavor(_flavor);

  static String get apiBaseUrl => config.apiBaseUrl;
  static String get appUrl => config.appUrl;
}
