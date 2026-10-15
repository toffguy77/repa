import 'package:flutter/material.dart';

import '../../theme/app_tokens.dart';

/// Shows a bottom sheet with the kit's surface, radius and scrim.
///
/// Exists so no screen has to restate the sheet's shape — the previous code repeated a
/// `BorderRadius.vertical(top: Radius.circular(24))` literal at every call site.
Future<T?> showAppSheet<T>({
  required BuildContext context,
  required WidgetBuilder builder,
  bool isScrollControlled = true,
}) {
  final t = context.t;
  return showModalBottomSheet<T>(
    context: context,
    isScrollControlled: isScrollControlled,
    backgroundColor: t.elevation.level2.surface,
    barrierColor: t.color.scrim,
    elevation: 0,
    shape: RoundedRectangleBorder(borderRadius: AppTokens.radius.sheet),
    builder: (context) => AppSheetBody(child: builder(context)),
  );
}

/// The grab handle plus padding every sheet shares.
class AppSheetBody extends StatelessWidget {
  const AppSheetBody({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    final t = context.t;
    return SafeArea(
      top: false,
      child: Padding(
        padding: EdgeInsets.all(AppTokens.space.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: t.color.border,
                borderRadius: BorderRadius.circular(AppTokens.radius.full),
              ),
            ),
            SizedBox(height: AppTokens.space.lg),
            child,
          ],
        ),
      ),
    );
  }
}
