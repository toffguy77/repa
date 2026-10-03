import 'package:flutter/material.dart';

import '../../theme/app_tokens.dart';

/// A pill-shaped progress bar. Pass a null [value] for the indeterminate state.
class AppProgressBar extends StatelessWidget {
  const AppProgressBar({
    super.key,
    this.value,
    this.height = 6,
    this.tone = AppProgressTone.accent,
  });

  final double? value;
  final double height;
  final AppProgressTone tone;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    final color = switch (tone) {
      AppProgressTone.accent => t.color.accent,
      AppProgressTone.energy => t.color.energy,
      AppProgressTone.success => t.color.success,
    };

    return ClipRRect(
      borderRadius: BorderRadius.circular(AppTokens.radius.full),
      child: LinearProgressIndicator(
        value: value,
        minHeight: height,
        backgroundColor: t.color.surface,
        color: color,
      ),
    );
  }
}

enum AppProgressTone { accent, energy, success }
