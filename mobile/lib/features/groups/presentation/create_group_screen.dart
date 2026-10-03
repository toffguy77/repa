import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../core/analytics/analytics_service.dart';
import 'groups_notifier.dart';
import '../../../core/theme/app_tokens.dart';
import 'widgets/invite_share_sheet.dart';
import '../../../core/widgets/kit/kit.dart';

const _allCategories = [
  ('HOT', '\u{1F525} Горячее'),
  ('FUNNY', '\u{1F602} Смешное'),
  ('SECRETS', '\u{1F92B} Секреты'),
  ('SKILLS', '\u{1F3AF} Навыки'),
  ('ROMANCE', '\u{1F496} Романтика'),
  ('STUDY', '\u{1F4DA} Учёба'),
];

class CreateGroupScreen extends ConsumerStatefulWidget {
  const CreateGroupScreen({super.key});

  @override
  ConsumerState<CreateGroupScreen> createState() => _CreateGroupScreenState();
}

class _CreateGroupScreenState extends ConsumerState<CreateGroupScreen> {
  final _nameController = TextEditingController();
  final _telegramController = TextEditingController();
  final _selectedCategories = <String>{};

  /// Null until the creator touches the switch, which is the whole point: the age-based default lives
  /// on the server, and sending a value we guessed would override it — wrongly, whenever the app does
  /// not know the creator's birth year. The switch below still *shows* the default so it is not a
  /// hidden setting; it just does not transmit it.
  bool? _kindOnlyChoice;

  @override
  void dispose() {
    _nameController.dispose();
    _telegramController.dispose();
    super.dispose();
  }

  bool get _isValid =>
      _nameController.text.trim().length >= 3 &&
      _selectedCategories.isNotEmpty;

  Future<void> _create() async {
    HapticFeedback.mediumImpact();
    final result = await ref.read(createGroupProvider.notifier).create(
          name: _nameController.text.trim(),
          categories: _selectedCategories.toList(),
          telegramUsername: _telegramController.text.trim().replaceAll('@', ''),
          kindOnly: _kindOnlyChoice,
        );
    if (result != null && mounted) {
      ref.read(analyticsProvider).logGroupCreated();
      ref.read(groupsListProvider.notifier).refresh();
      _showInviteSheet(result.group.inviteCode);
    }
  }

  void _showInviteSheet(String inviteCode) {
    showAppSheet<void>(
      context: context,
      builder: (ctx) => InviteShareSheet(
        inviteCode: inviteCode,
        title: 'Группа создана!',
        onDone: () {
          Navigator.pop(ctx);
          context.go('/home');
        },
      ),
    );
  }

  List<(String, String)> _availableCategories(WidgetRef ref) {
    if (_isUnder18(ref)) {
      return _allCategories.where((c) => c.$1 != 'ROMANCE').toList();
    }
    return _allCategories;
  }

  /// An unknown birth year counts as under 18, the same way the ROMANCE restriction treats it and the
  /// same way the server's default does — one notion of "under 18" across the product.
  bool _isUnder18(WidgetRef ref) {
    final birthYear = ref.read(authProvider).user?.birthYear;
    if (birthYear == null) return true;
    return DateTime.now().year - birthYear < 18;
  }

  /// What the switch shows before anyone touches it. Mirrors the server's rule so the position the
  /// creator sees is the one that will actually apply.
  bool _kindOnlyDisplayed(WidgetRef ref) => _kindOnlyChoice ?? _isUnder18(ref);

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(createGroupProvider);
    final categories = _availableCategories(ref);

    return Scaffold(
      appBar: AppBar(title: const Text('Новая группа')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text('Название', style: context.ts.body),
          const SizedBox(height: 8),
          TextField(
            controller: _nameController,
            maxLength: 40,
            decoration: const InputDecoration(
              hintText: 'Название группы',
              counterText: '',
            ),
            onChanged: (_) => setState(() {}),
          ),
          const SizedBox(height: 20),
          Text('Категории вопросов', style: context.ts.body),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: categories.map((cat) {
              final selected = _selectedCategories.contains(cat.$1);
              return FilterChip(
                label: Text(cat.$2),
                selected: selected,
                selectedColor: context.t.color.accentFill.withValues(alpha: 0.18),
                checkmarkColor: context.t.color.accent,
                onSelected: (val) {
                  HapticFeedback.selectionClick();
                  setState(() {
                    if (val) {
                      _selectedCategories.add(cat.$1);
                    } else {
                      _selectedCategories.remove(cat.$1);
                    }
                  });
                },
              );
            }).toList(),
          ),
          const SizedBox(height: 20),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text('Только добрые вопросы',
                          style: context.ts.body),
                    ),
                    Switch(
                      value: _kindOnlyDisplayed(ref),
                      onChanged: (val) {
                        HapticFeedback.selectionClick();
                        setState(() => _kindOnlyChoice = val);
                      },
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                // Stated as what the group will and will not be asked, rather than as an
                // unexplained toggle — nobody turns on a switch they cannot predict.
                Text(
                  'Группа будет получать только приятные и нейтральные вопросы. '
                  'Колкие вопросы — про сплетни, зависть, «кто хуже всех» — приходить не будут.',
                  style: context.ts.caption,
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          Text('Telegram (необязательно)', style: context.ts.body),
          const SizedBox(height: 8),
          TextField(
            controller: _telegramController,
            decoration: const InputDecoration(
              hintText: '@username',
              prefixIcon: Icon(Icons.telegram),
            ),
          ),
          if (state.error != null) ...[
            const SizedBox(height: 12),
            Text(
              state.error!,
              style: context.ts.caption.copyWith(color: context.t.color.danger),
            ),
          ],
          const SizedBox(height: 32),
          SizedBox(
            height: 50,
            child: ElevatedButton(
              onPressed: _isValid && !state.loading ? _create : null,
              child: state.loading
                  ? SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: context.t.color.onAccentFill,
                      ),
                    )
                  : const Text('Создать'),
            ),
          ),
        ],
      ),
    );
  }
}
