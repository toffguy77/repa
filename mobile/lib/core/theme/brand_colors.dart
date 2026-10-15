import 'package:flutter/material.dart';

/// Colours owned by someone else.
///
/// These are not design tokens: they do not change with the theme and they must not be
/// "adjusted to fit", because they identify a third-party product. They live here so a
/// screen never carries a bare hex literal.
class BrandColors {
  const BrandColors._();

  /// Telegram's brand blue, used on the button that opens Telegram.
  static const telegram = Color(0xFF2AABEE);
}
