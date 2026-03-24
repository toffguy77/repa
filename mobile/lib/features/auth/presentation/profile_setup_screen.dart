import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import 'auth_notifier.dart';

const _avatarEmojis = [
  '\u{1F60E}', '\u{1F47B}', '\u{1F525}', '\u{1F680}', '\u{1F308}',
  '\u{1F3AE}', '\u{1F3B5}', '\u{1F4A1}', '\u{1F40D}', '\u{1F43B}',
  '\u{1F431}', '\u{1F436}', '\u{1F984}', '\u{1F47E}', '\u{1F916}',
  '\u{1F30A}', '\u{26A1}', '\u{1F48E}', '\u{1F3AF}', '\u{1F389}',
];

class ProfileSetupScreen extends ConsumerStatefulWidget {
  const ProfileSetupScreen({super.key});

  @override
  ConsumerState<ProfileSetupScreen> createState() =>
      _ProfileSetupScreenState();
}

class _ProfileSetupScreenState extends ConsumerState<ProfileSetupScreen> {
  final _usernameController = TextEditingController();
  final _birthYearController = TextEditingController();
  String _selectedEmoji = _avatarEmojis[0];
  Timer? _debounce;

  @override
  void dispose() {
    _debounce?.cancel();
    _usernameController.dispose();
    _birthYearController.dispose();
    super.dispose();
  }

  void _onUsernameChanged(String value) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 500), () {
      ref.read(profileSetupProvider.notifier).checkUsername(value);
    });
  }

  bool get _isValid {
    final username = _usernameController.text.trim();
    final birthYearText = _birthYearController.text.trim();
    if (username.length < 3) return false;
    if (birthYearText.isEmpty) return false;
    final birthYear = int.tryParse(birthYearText);
    final maxYear = DateTime.now().year - 14;
    final minYear = DateTime.now().year - 100;
    if (birthYear == null || birthYear < minYear || birthYear > maxYear) return false;
    return true;
  }

  Future<void> _submit() async {
    if (!_isValid) return;
    final success = await ref.read(profileSetupProvider.notifier).submit(
          username: _usernameController.text.trim(),
          avatarEmoji: _selectedEmoji,
          birthYear: int.parse(_birthYearController.text.trim()),
        );
    if (!success && mounted) {
      // Error is shown via state
    }
    // Navigation handled by router redirect
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(profileSetupProvider);

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 60),
              Text('Расскажи о себе', style: AppTextStyles.headline1),
              const SizedBox(height: 8),
              Text(
                'Заполни профиль, чтобы друзья тебя узнали',
                style: AppTextStyles.bodySecondary,
              ),
              const SizedBox(height: 32),

              // Avatar emoji picker
              Text('Выбери аватар', style: AppTextStyles.body),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: _avatarEmojis.map((emoji) {
                  final selected = emoji == _selectedEmoji;
                  return GestureDetector(
                    onTap: () => setState(() => _selectedEmoji = emoji),
                    child: Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        color:
                            selected ? AppColors.primaryLight : AppColors.surface,
                        borderRadius: BorderRadius.circular(12),
                        border: selected
                            ? Border.all(color: AppColors.primary, width: 2)
                            : null,
                      ),
                      child: Center(
                        child: Text(emoji, style: const TextStyle(fontSize: 24)),
                      ),
                    ),
                  );
                }).toList(),
              ),
              const SizedBox(height: 24),

              // Username
              Text('Имя пользователя', style: AppTextStyles.body),
              const SizedBox(height: 8),
              TextField(
                controller: _usernameController,
                style: AppTextStyles.body,
                decoration: InputDecoration(
                  hintText: 'Минимум 3 символа',
                  suffixIcon: _usernameSuffix(state),
                ),
                onChanged: (v) {
                  setState(() {});
                  _onUsernameChanged(v);
                },
              ),
              if (state.usernameAvailable == false) ...[
                const SizedBox(height: 4),
                Text(
                  'Имя занято',
                  style: TextStyle(color: AppColors.error, fontSize: 13),
                ),
              ],
              const SizedBox(height: 24),

              // Birth year
              Text('Год рождения', style: AppTextStyles.body),
              const SizedBox(height: 8),
              TextField(
                controller: _birthYearController,
                keyboardType: TextInputType.number,
                style: AppTextStyles.body,
                inputFormatters: [
                  FilteringTextInputFormatter.digitsOnly,
                  LengthLimitingTextInputFormatter(4),
                ],
                decoration: const InputDecoration(hintText: 'Например, 2005'),
                onChanged: (_) => setState(() {}),
              ),
              const SizedBox(height: 8),
              Text(
                'Нужен для подбора контента по возрасту',
                style: AppTextStyles.caption,
              ),

              if (state.error != null) ...[
                const SizedBox(height: 16),
                Text(state.error!, style: TextStyle(color: AppColors.error)),
              ],
              const SizedBox(height: 32),

              ElevatedButton(
                onPressed: _isValid && !state.loading ? _submit : null,
                child: state.loading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Text('Готово'),
              ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }

  Widget? _usernameSuffix(ProfileSetupState state) {
    if (state.checkingUsername) {
      return const Padding(
        padding: EdgeInsets.all(12),
        child: SizedBox(
          width: 20,
          height: 20,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
      );
    }
    if (state.usernameAvailable == true) {
      return const Icon(Icons.check_circle, color: AppColors.success);
    }
    if (state.usernameAvailable == false) {
      return const Icon(Icons.cancel, color: AppColors.error);
    }
    return null;
  }
}
