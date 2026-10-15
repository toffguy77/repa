import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_tokens.dart';
import '../../../core/widgets/error_state_widget.dart';
import '../domain/chronicle.dart';
import 'chronicle_notifier.dart';
import 'widgets/chronicle_season_card.dart';
import 'widgets/guess_standing_card.dart';

/// The group's shared record of its past seasons.
class ChronicleScreen extends ConsumerStatefulWidget {
  const ChronicleScreen({super.key, required this.groupId});

  final String groupId;

  @override
  ConsumerState<ChronicleScreen> createState() => _ChronicleScreenState();
}

class _ChronicleScreenState extends ConsumerState<ChronicleScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
        () => ref.read(chronicleProvider(widget.groupId).notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(chronicleProvider(widget.groupId));

    return Scaffold(
      appBar: AppBar(title: const Text('История группы')),
      body: _body(state),
    );
  }

  Widget _body(ChronicleState state) {
    if (state.error != null && state.chronicle == null) {
      return ErrorStateWidget(
        message: state.error!,
        onRetry: () => ref.read(chronicleProvider(widget.groupId).notifier).load(),
      );
    }
    final chronicle = state.chronicle;
    if (chronicle == null) {
      return const Center(child: CircularProgressIndicator());
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(chronicleProvider(widget.groupId).notifier).load(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          GuessStandingCard(standing: chronicle.standing),
          const SizedBox(height: 12),
          if (chronicle.isEmpty)
            _EmptyChronicle()
          else ...[
            for (final season in chronicle.seasons) ...[
              ChronicleSeasonCard(season: season),
              const SizedBox(height: 12),
            ],
            if (chronicle.isTruncated)
              Text(
                'Показаны последние ${chronicle.seasonsShown} из ${chronicle.seasonsTotal}',
                style: context.ts.caption,
                textAlign: TextAlign.center,
              ),
          ],
        ],
      ),
    );
  }
}

class _EmptyChronicle extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 48),
      child: Column(
        children: [
          Text('\u{1F4D6}', style: context.ts.display2),
          const SizedBox(height: 12),
          Text('Пока ничего не произошло', style: context.ts.body),
          const SizedBox(height: 4),
          // Names what will fill it, rather than just reporting that it is empty.
          Text(
            'После первой репы здесь останется, кто что получил — и так каждую пятницу.',
            style: context.ts.caption,
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
