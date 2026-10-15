import 'package:flutter/material.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../domain/chronicle.dart';

/// One past season in the group's record.
class ChronicleSeasonCard extends StatelessWidget {
  const ChronicleSeasonCard({super.key, required this.season});

  final ChronicleSeason season;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text('Репа №${season.number}', style: context.ts.body),
              ),
              Text(_shortDate(season.revealAt), style: context.ts.caption),
            ],
          ),
          const SizedBox(height: 8),
          if (season.entries.isEmpty)
            // A season can be present with no entries: every standout result in it belongs to someone
            // this reader has blocked. The season still happened, so it stays — removing it would tell
            // the reader that a block is in play.
            Text('Нечего показать', style: context.ts.caption)
          else
            ...season.entries.map((e) => _EntryRow(entry: e)),
        ],
      ),
    );
  }
}

class _EntryRow extends StatelessWidget {
  const _EntryRow({required this.entry});

  final ChronicleEntry entry;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(entry.questionText, style: context.ts.caption),
          const SizedBox(height: 2),
          Row(
            children: [
              Text(entry.avatarEmoji ?? '\u{1F346}'),
              const SizedBox(width: 6),
              Expanded(
                child: Text(entry.username, style: context.ts.body),
              ),
              Text('${entry.percentage.round()}%',
                  style: context.ts.body
                      .copyWith(color: context.t.color.accent)),
            ],
          ),
          if (entry.smallSample) ...[
            const SizedBox(height: 2),
            // The same caveat the live results carry. Attached per entry rather than per season,
            // because it is a property of how many people voted on that result.
            Text(
              'Голосовало ${entry.totalVoters} — в таком составе проценты выдают, кто как голосовал',
              style: context.ts.caption.copyWith(color: context.t.color.warning),
            ),
          ],
        ],
      ),
    );
  }
}

/// Renders `2026-09-18T17:00:00Z` as `18.09`. Falls back to the raw string rather than throwing: a
/// timestamp the client cannot parse is not a reason to lose the season.
String _shortDate(String iso) {
  final parsed = DateTime.tryParse(iso);
  if (parsed == null) return iso;
  final local = parsed.toLocal();
  return '${local.day.toString().padLeft(2, '0')}.${local.month.toString().padLeft(2, '0')}';
}
