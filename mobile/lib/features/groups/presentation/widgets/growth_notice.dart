import 'package:flutter/material.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../domain/group.dart';

/// What inviting more people would change.
///
/// Shown only while there is a real threshold left. Above the last one this renders nothing: a group
/// that is big enough is told so by the absence of a goal, not by a line inviting it to keep growing.
/// The number and which consequence both come from the server (`next_threshold`, derived from
/// `internal/eligibility`), so a rule change cannot leave this copy behind.
class GrowthNotice extends StatelessWidget {
  const GrowthNotice({super.key, required this.threshold});

  final GrowthThreshold? threshold;

  @override
  Widget build(BuildContext context) {
    final t = threshold;
    if (t == null) return const SizedBox.shrink();

    return AppCard(
      child: Row(
        children: [
          const Text('\u{1F465}'),
          const SizedBox(width: 8),
          Expanded(
            child: Text(_copy(t), style: context.ts.caption),
          ),
        ],
      ),
    );
  }

  static String _copy(GrowthThreshold t) {
    final people = _plural(t.needed);
    switch (t.unlocks) {
      case GrowthUnlock.reveal:
        return 'Позови ещё $people — без этого репа не откроется.';
      case GrowthUnlock.detector:
        return 'Позови ещё $people — откроется детектор, и проценты перестанут выдавать, кто голосовал.';
      case GrowthUnlock.unknown:
        // A threshold this client does not know about: state the number without claiming what it
        // does, rather than dropping the line or inventing a consequence.
        return 'Позови ещё $people — в группе побольше открывается больше.';
    }
  }

  static String _plural(int n) {
    final mod100 = n % 100;
    if (mod100 >= 11 && mod100 <= 14) return '$n человек';
    switch (n % 10) {
      case 1:
        return '$n человека';
      case 2:
      case 3:
      case 4:
        return '$n человека';
      default:
        return '$n человек';
    }
  }
}
