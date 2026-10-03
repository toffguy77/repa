import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../theme/app_tokens.dart';

/// A small selectable tag. Drawn compact, but its touch target is still the minimum.
class AppChip extends StatelessWidget {
  const AppChip({
    super.key,
    required this.label,
    this.selected = false,
    this.onTap,
    this.leading,
    this.tone = AppChipTone.neutral,
  });

  final String label;
  final bool selected;
  final VoidCallback? onTap;
  final Widget? leading;
  final AppChipTone tone;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final radius = AppTokens.radius.chip;

    final Color background;
    final Color foreground;
    final Color outline;
    switch (tone) {
      case AppChipTone.neutral:
        background = selected ? t.color.accentFill : t.color.surface;
        foreground = selected ? t.color.onAccentFill : t.color.textPrimary;
        outline = selected ? t.color.accentFill : t.color.border;
      case AppChipTone.energy:
        background = selected ? t.color.energyFill : t.color.surface;
        foreground = selected ? t.color.onEnergyFill : t.color.energy;
        outline = selected ? t.color.energyFill : t.color.border;
    }

    final visual = DecoratedBox(
      decoration: BoxDecoration(
        color: background,
        borderRadius: radius,
        border: Border.all(color: outline, width: AppTokens.border.hairline),
      ),
      child: Padding(
        padding: EdgeInsets.symmetric(
          horizontal: AppTokens.space.md,
          vertical: AppTokens.space.sm,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (leading != null) ...[
              leading!,
              SizedBox(width: AppTokens.space.xs),
            ],
            Text(
              label,
              style: AppTokens.text.caption.copyWith(
                color: foreground,
                fontWeight: selected ? FontWeight.w600 : FontWeight.w400,
              ),
            ),
          ],
        ),
      ),
    );

    if (onTap == null) return visual;

    // The chip is drawn small; the target is not.
    return ConstrainedBox(
      constraints: BoxConstraints(minHeight: AppTokens.border.minTouchTarget),
      child: Material(
        color: Colors.transparent,
        borderRadius: radius,
        child: InkWell(
          borderRadius: radius,
          onTap: () {
            HapticFeedback.selectionClick();
            onTap!.call();
          },
          child: Center(widthFactor: 1, child: visual),
        ),
      ),
    );
  }
}

enum AppChipTone { neutral, energy }
