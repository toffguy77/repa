import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:share_plus/share_plus.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import 'groups_notifier.dart';

const _categories = [
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
        );
    if (result != null && mounted) {
      ref.read(groupsListProvider.notifier).refresh();
      _showInviteSheet(result.inviteUrl);
    }
  }

  void _showInviteSheet(String inviteUrl) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('\u{1F389}', style: TextStyle(fontSize: 48)),
            const SizedBox(height: 12),
            Text('Группа создана!', style: AppTextStyles.headline2),
            const SizedBox(height: 8),
            Text(
              'Поделись ссылкой с друзьями',
              style: AppTextStyles.bodySecondary,
            ),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      inviteUrl,
                      style: AppTextStyles.caption,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.copy, size: 20),
                    onPressed: () {
                      Clipboard.setData(ClipboardData(text: inviteUrl));
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Ссылка скопирована')),
                      );
                    },
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              height: 50,
              child: ElevatedButton.icon(
                onPressed: () {
                  Share.share(inviteUrl);
                },
                icon: const Icon(Icons.share),
                label: const Text('Поделиться'),
              ),
            ),
            const SizedBox(height: 12),
            TextButton(
              onPressed: () {
                Navigator.pop(ctx);
                context.go('/home');
              },
              child: const Text('Готово'),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(createGroupProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Новая группа')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text('Название', style: AppTextStyles.body),
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
          Text('Категории вопросов', style: AppTextStyles.body),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _categories.map((cat) {
              final selected = _selectedCategories.contains(cat.$1);
              return FilterChip(
                label: Text(cat.$2),
                selected: selected,
                selectedColor: AppColors.primaryLight,
                checkmarkColor: AppColors.primary,
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
          Text('Telegram (необязательно)', style: AppTextStyles.body),
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
              style: AppTextStyles.caption.copyWith(color: AppColors.error),
            ),
          ],
          const SizedBox(height: 32),
          SizedBox(
            height: 50,
            child: ElevatedButton(
              onPressed: _isValid && !state.loading ? _create : null,
              child: state.loading
                  ? const SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
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
