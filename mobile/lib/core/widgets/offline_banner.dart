import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/connectivity_provider.dart';
import '../theme/app_tokens.dart';

class OfflineBanner extends ConsumerWidget {
  final Widget child;

  const OfflineBanner({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isOnline = ref.watch(connectivityProvider);
    final t = context.t;

    return Column(
      children: [
        AnimatedSize(
          duration: context.motion(MotionClass.surface),
          curve: context.motionCurve(MotionClass.surface),
          child: isOnline
              ? const SizedBox.shrink()
              : Container(
                  width: double.infinity,
                  color: t.color.danger,
                  padding: EdgeInsets.only(
                    top: MediaQuery.of(context).padding.top + AppTokens.space.xs,
                    bottom: AppTokens.space.xs,
                  ),
                  child: Text(
                    'Нет соединения',
                    textAlign: TextAlign.center,
                    style: AppTokens.text.caption.copyWith(
                      color: t.color.onAccentFill,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
        ),
        Expanded(child: child),
      ],
    );
  }
}
