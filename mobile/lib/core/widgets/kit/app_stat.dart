import 'package:flutter/material.dart';

import '../../theme/app_tokens.dart';

/// A large numeric readout with a micro-label under it. Uses tabular figures so a changing
/// value does not shift the layout around it.
class AppStat extends StatelessWidget {
  const AppStat({
    super.key,
    required this.value,
    required this.label,
    this.tone = AppStatTone.primary,
    this.compact = false,
  });

  final String value;
  final String label;
  final AppStatTone tone;

  /// Use the numeric body role instead of the display role — for stats inside a dense row.
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final color = switch (tone) {
      AppStatTone.primary => t.color.textPrimary,
      AppStatTone.accent => t.color.accent,
      AppStatTone.energy => t.color.energy,
    };

    final valueStyle = compact
        ? AppTokens.text.numeric
        : AppTokens.text.display2.copyWith(
            fontFeatures: AppTokens.text.numeric.fontFeatures,
          );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(value, style: valueStyle.copyWith(color: color)),
        SizedBox(height: AppTokens.space.xs),
        Text(
          label.toUpperCase(),
          style: AppTokens.text.label.copyWith(color: t.color.textSecondary),
        ),
      ],
    );
  }
}

enum AppStatTone { primary, accent, energy }
