import 'package:flutter/material.dart';

import '../../theme/app_tokens.dart';

/// A panel. Takes its surface, outline and shadows from the active theme's level-1
/// elevation, which is how dark mode gets depth without shadows.
class AppCard extends StatelessWidget {
  const AppCard({
    super.key,
    required this.child,
    this.onTap,
    this.padding,
    this.raised = false,
  });

  final Widget child;
  final VoidCallback? onTap;
  final EdgeInsetsGeometry? padding;

  /// Use the level-2 surface — for a panel sitting on another panel.
  final bool raised;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final level = raised ? t.elevation.level2 : t.elevation.level1;
    final radius = AppTokens.radius.card;

    final content = Padding(
      padding: padding ?? EdgeInsets.all(AppTokens.space.lg),
      child: child,
    );

    return DecoratedBox(
      decoration: BoxDecoration(
        color: level.surface,
        borderRadius: radius,
        border: Border.all(color: level.outline, width: AppTokens.border.hairline),
        boxShadow: level.shadows,
      ),
      child: onTap == null
          ? content
          : Material(
              color: Colors.transparent,
              borderRadius: radius,
              child: InkWell(
                onTap: onTap,
                borderRadius: radius,
                child: content,
              ),
            ),
    );
  }
}

/// A section title, optionally with a trailing action and an uppercase overline.
class AppSectionHeader extends StatelessWidget {
  const AppSectionHeader({
    super.key,
    required this.title,
    this.overline,
    this.trailing,
  });

  final String title;
  final String? overline;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    return Padding(
      padding: EdgeInsets.only(bottom: AppTokens.space.md),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (overline != null)
                  Text(
                    overline!.toUpperCase(),
                    style: AppTokens.text.label.copyWith(color: t.color.textSecondary),
                  ),
                Text(
                  title,
                  style: AppTokens.text.heading2.copyWith(color: t.color.textPrimary),
                ),
              ],
            ),
          ),
          // ignore: use_null_aware_elements
          if (trailing != null) trailing!,
        ],
      ),
    );
  }
}
