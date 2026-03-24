import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import 'settings_notifier.dart';

final _packageInfoProvider = FutureProvider<PackageInfo>((ref) {
  return PackageInfo.fromPlatform();
});

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(settingsProvider);
    final user = state.user;

    return Scaffold(
      appBar: AppBar(title: const Text('Настройки')),
      body: ListView(
        padding: const EdgeInsets.symmetric(vertical: 8),
        children: [
          // --- Profile section ---
          _SectionHeader('Профиль'),
          _AvatarTile(
            avatarUrl: user?.avatarUrl,
            avatarEmoji: user?.avatarEmoji,
            loading: state.savingAvatar,
            onTap: () => _showAvatarPicker(context, ref),
          ),
          ListTile(
            leading: const Icon(Icons.alternate_email),
            title: const Text('Никнейм'),
            subtitle: Text(user?.username ?? ''),
          ),
          ListTile(
            leading: const Icon(Icons.cake_outlined),
            title: const Text('Год рождения'),
            subtitle: Text(user?.birthYear?.toString() ?? 'Не указан'),
          ),
          const Divider(height: 32),

          // --- Push preferences ---
          _SectionHeader('Уведомления'),
          ...List.generate(state.pushPrefs.length, (i) {
            final pref = state.pushPrefs[i];
            return SwitchListTile(
              title: Text(pref.label),
              value: pref.enabled,
              activeTrackColor: AppColors.primary,
              onChanged: (v) =>
                  ref.read(settingsProvider.notifier).togglePushPref(i, v),
            );
          }),
          const Divider(height: 32),

          // --- Account ---
          _SectionHeader('Аккаунт'),
          ListTile(
            leading: Icon(Icons.logout, color: AppColors.error),
            title: Text('Выйти', style: TextStyle(color: AppColors.error)),
            onTap: () => _confirmLogout(context, ref),
          ),
          ListTile(
            leading: Icon(Icons.delete_forever, color: AppColors.error),
            title: Text('Удалить аккаунт',
                style: TextStyle(color: AppColors.error)),
            onTap: () => _confirmDelete(context, ref),
          ),
          const Divider(height: 32),

          // --- About ---
          _SectionHeader('О приложении'),
          Consumer(builder: (context, ref, _) {
            final info = ref.watch(_packageInfoProvider);
            final version = info.valueOrNull?.version ?? '...';
            final build = info.valueOrNull?.buildNumber ?? '';
            return ListTile(
              leading: const Icon(Icons.info_outline),
              title: const Text('Версия'),
              subtitle: Text('$version ($build)'),
            );
          }),
          ListTile(
            leading: const Icon(Icons.privacy_tip_outlined),
            title: const Text('Политика конфиденциальности'),
            onTap: () => launchUrl(
              Uri.parse('https://repa.app/privacy'),
              mode: LaunchMode.externalApplication,
            ),
          ),
          ListTile(
            leading: const Icon(Icons.description_outlined),
            title: const Text('Правила использования'),
            onTap: () => launchUrl(
              Uri.parse('https://repa.app/terms'),
              mode: LaunchMode.externalApplication,
            ),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }

  void _showAvatarPicker(BuildContext context, WidgetRef ref) {
    HapticFeedback.lightImpact();
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.camera_alt),
              title: const Text('Камера'),
              onTap: () {
                Navigator.pop(context);
                ref
                    .read(settingsProvider.notifier)
                    .pickAndUploadAvatar(ImageSource.camera);
              },
            ),
            ListTile(
              leading: const Icon(Icons.photo_library),
              title: const Text('Галерея'),
              onTap: () {
                Navigator.pop(context);
                ref
                    .read(settingsProvider.notifier)
                    .pickAndUploadAvatar(ImageSource.gallery);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _confirmLogout(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Выйти из аккаунта?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              ref.read(settingsProvider.notifier).logout();
            },
            child: Text('Выйти', style: TextStyle(color: AppColors.error)),
          ),
        ],
      ),
    );
  }

  void _confirmDelete(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Удалить аккаунт?'),
        content: const Text(
          'Все данные будут удалены безвозвратно. Это действие нельзя отменить.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Отмена'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              ref.read(settingsProvider.notifier).deleteAccount();
            },
            child: Text('Удалить', style: TextStyle(color: AppColors.error)),
          ),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;
  const _SectionHeader(this.title);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
      child: Text(
        title,
        style: AppTextStyles.caption.copyWith(fontWeight: FontWeight.w600),
      ),
    );
  }
}

class _AvatarTile extends StatelessWidget {
  final String? avatarUrl;
  final String? avatarEmoji;
  final bool loading;
  final VoidCallback onTap;

  const _AvatarTile({
    this.avatarUrl,
    this.avatarEmoji,
    required this.loading,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Stack(
        children: [
          CircleAvatar(
            radius: 28,
            backgroundColor: AppColors.primaryLight,
            backgroundImage:
                avatarUrl != null ? NetworkImage(avatarUrl!) : null,
            child: avatarUrl == null
                ? Text(
                    avatarEmoji ?? '🍆',
                    style: const TextStyle(fontSize: 28),
                  )
                : null,
          ),
          if (loading)
            const Positioned.fill(
              child: Center(
                child: SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ),
            ),
        ],
      ),
      title: const Text('Фото профиля'),
      subtitle: const Text('Нажмите, чтобы изменить'),
      onTap: loading ? null : onTap,
    );
  }
}
