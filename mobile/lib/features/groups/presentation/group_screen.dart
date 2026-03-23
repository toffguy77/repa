import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import 'groups_notifier.dart';
import 'widgets/member_avatar.dart';

class GroupScreen extends ConsumerStatefulWidget {
  final String groupId;

  const GroupScreen({super.key, required this.groupId});

  @override
  ConsumerState<GroupScreen> createState() => _GroupScreenState();
}

class _GroupScreenState extends ConsumerState<GroupScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(groupDetailProvider(widget.groupId).notifier).load();
    });
  }

  void _shareInvite() {
    final detail = ref.read(groupDetailProvider(widget.groupId)).detail;
    if (detail == null) return;
    final url = 'https://repa.app/join/${detail.group.inviteCode}';
    Share.share(url);
  }

  void _openTelegram(String username) {
    launchUrl(Uri.parse('https://t.me/$username'));
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(groupDetailProvider(widget.groupId));

    if (state.loading && state.detail == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (state.error != null && state.detail == null) {
      return Scaffold(
        appBar: AppBar(),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(state.error!, style: AppTextStyles.body),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => ref
                    .read(groupDetailProvider(widget.groupId).notifier)
                    .load(),
                child: const Text('Повторить'),
              ),
            ],
          ),
        ),
      );
    }

    final detail = state.detail;
    if (detail == null) return const SizedBox.shrink();

    final group = detail.group;
    final season = detail.activeSeason;
    final members = detail.members;
    final currentUserId = ref.watch(authProvider).user?.id;
    final isAdmin = currentUserId == group.adminId;

    return Scaffold(
      appBar: AppBar(
        title: Text(group.name),
        actions: [
          IconButton(
            icon: const Icon(Icons.share),
            onPressed: _shareInvite,
            tooltip: 'Пригласить',
          ),
          if (group.telegramUsername != null)
            IconButton(
              icon: const Icon(Icons.telegram),
              onPressed: () => _openTelegram(group.telegramUsername!),
              tooltip: 'Telegram',
            ),
          if (isAdmin)
            IconButton(
              icon: const Icon(Icons.settings_outlined),
              onPressed: () => context.push('/groups/${widget.groupId}/telegram'),
              tooltip: 'Настройки Telegram',
            ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () =>
            ref.read(groupDetailProvider(widget.groupId).notifier).load(),
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Season status
            if (season != null) ...[
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.05),
                      blurRadius: 10,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Сезон',
                      style: AppTextStyles.headline2.copyWith(fontSize: 18),
                    ),
                    const SizedBox(height: 12),
                    if (season.status == 'VOTING') ...[
                      ClipRRect(
                        borderRadius: BorderRadius.circular(4),
                        child: LinearProgressIndicator(
                          value: season.totalCount > 0
                              ? season.votedCount / season.totalCount
                              : 0,
                          backgroundColor: AppColors.surface,
                          color: AppColors.primary,
                          minHeight: 8,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '${season.votedCount} из ${season.totalCount} проголосовали',
                        style: AppTextStyles.caption,
                      ),
                      const SizedBox(height: 12),
                      SizedBox(
                        width: double.infinity,
                        height: 48,
                        child: ElevatedButton(
                          onPressed: season.userVoted
                              ? null
                              : () {
                                  HapticFeedback.mediumImpact();
                                  context.go(
                                      '/groups/${widget.groupId}/vote/${season.id}');
                                },
                          child: Text(
                            season.userVoted
                                ? 'Ждём пятницы'
                                : 'Проголосовать',
                          ),
                        ),
                      ),
                    ],
                    if (season.status == 'REVEALED') ...[
                      Text(
                        'Результаты готовы!',
                        style: AppTextStyles.body.copyWith(
                          color: AppColors.success,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: 12),
                      SizedBox(
                        width: double.infinity,
                        height: 48,
                        child: ElevatedButton(
                          onPressed: () {
                            HapticFeedback.mediumImpact();
                            context.push(
                              '/groups/${widget.groupId}/reveal/${season.id}?status=REVEALED',
                            );
                          },
                          child: const Text('Открыть репу'),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 20),
            ],

            // Members
            Text('Участники', style: AppTextStyles.headline2),
            const SizedBox(height: 12),
            ...members.map(
              (m) => Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: InkWell(
                  borderRadius: BorderRadius.circular(8),
                  onTap: () {
                    HapticFeedback.lightImpact();
                    context.push('/groups/${widget.groupId}/members/${m.id}');
                  },
                  child: Row(
                    children: [
                      MemberAvatar(
                        avatarEmoji: m.avatarEmoji,
                        avatarUrl: m.avatarUrl,
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(m.username, style: AppTextStyles.body),
                      ),
                      if (m.isAdmin)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: AppColors.primaryLight,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            'Админ',
                            style: AppTextStyles.caption.copyWith(
                              color: AppColors.primary,
                              fontSize: 12,
                            ),
                          ),
                        ),
                      const Icon(Icons.chevron_right, color: AppColors.textSecondary),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
