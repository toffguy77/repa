import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../theme/app_tokens.dart';

enum AppButtonVariant {
  /// The one action a screen wants you to take.
  primary,

  /// A real alternative, outlined rather than filled.
  secondary,

  /// Low-emphasis, no container.
  ghost,
}

/// The app's button. Wraps the Material button for its variant so semantics, focus and ink
/// keep working, and binds tokens plus the 44px minimum target.
class AppButton extends StatelessWidget {
  const AppButton({
    super.key,
    required this.label,
    this.onPressed,
    this.variant = AppButtonVariant.primary,
    this.icon,
    this.loading = false,
    this.expand = true,
    this.danger = false,
  });

  final String label;
  final VoidCallback? onPressed;
  final AppButtonVariant variant;
  final IconData? icon;

  /// Shows a spinner and blocks taps. Separate from a null [onPressed] so a screen can
  /// distinguish "working" from "not available".
  final bool loading;

  /// Fill the available width. Off for buttons sitting in a row.
  final bool expand;

  /// Recolours a primary button for a destructive action.
  final bool danger;

  bool get _enabled => onPressed != null && !loading;

  void _handleTap() {
    HapticFeedback.lightImpact();
    onPressed!.call();
  }

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final child = _content(context);
    final minSize = Size(
      expand ? double.infinity : 0,
      AppTokens.border.minTouchTarget + 4,
    );

    final Widget button;
    switch (variant) {
      case AppButtonVariant.primary:
        button = ElevatedButton(
          onPressed: _enabled ? _handleTap : null,
          style: ElevatedButton.styleFrom(
            backgroundColor: danger ? t.color.danger : t.color.accentFill,
            foregroundColor: t.color.onAccentFill,
            disabledBackgroundColor:
                (danger ? t.color.danger : t.color.accentFill).withValues(alpha: 0.35),
            disabledForegroundColor: t.color.onAccentFill.withValues(alpha: 0.6),
            minimumSize: minSize,
            elevation: 0,
            shape: RoundedRectangleBorder(borderRadius: AppTokens.radius.control),
            textStyle: AppTokens.text.bodyStrong,
          ),
          child: child,
        );
      case AppButtonVariant.secondary:
        button = OutlinedButton(
          onPressed: _enabled ? _handleTap : null,
          style: OutlinedButton.styleFrom(
            foregroundColor: t.color.textPrimary,
            disabledForegroundColor: t.color.textSecondary,
            minimumSize: minSize,
            side: BorderSide(color: t.color.border, width: AppTokens.border.hairline),
            shape: RoundedRectangleBorder(borderRadius: AppTokens.radius.control),
            textStyle: AppTokens.text.bodyStrong,
          ),
          child: child,
        );
      case AppButtonVariant.ghost:
        button = TextButton(
          onPressed: _enabled ? _handleTap : null,
          style: TextButton.styleFrom(
            foregroundColor: danger ? t.color.danger : t.color.accent,
            disabledForegroundColor: t.color.textSecondary,
            minimumSize: Size(expand ? double.infinity : 0, AppTokens.border.minTouchTarget),
            shape: RoundedRectangleBorder(borderRadius: AppTokens.radius.control),
            textStyle: AppTokens.text.bodyStrong,
          ),
          child: child,
        );
    }

    return button;
  }

  Widget _content(BuildContext context) {
    if (loading) {
      final color = variant == AppButtonVariant.primary
          ? context.t.color.onAccentFill
          : context.t.color.accent;
      return SizedBox(
        height: 18,
        width: 18,
        child: CircularProgressIndicator(strokeWidth: 2, color: color),
      );
    }
    if (icon == null) return Text(label);
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 18),
        SizedBox(width: AppTokens.space.sm),
        Text(label),
      ],
    );
  }
}
