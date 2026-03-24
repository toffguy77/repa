import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../core/providers/connectivity_provider.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_text_styles.dart';
import '../../../core/widgets/empty_state_widget.dart';
import '../../../core/widgets/error_state_widget.dart';
import '../../../core/widgets/reveal_countdown_widget.dart';
import '../../../core/widgets/skeleton_loader.dart';
import '../../question_vote/presentation/question_vote_notifier.dart';
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

    // Auto-refresh on reconnect
    ref.listen<bool>(connectivityProvider, (prev, next) {
      if (prev == false && next == true) {
        ref.read(groupDetailProvider(widget.groupId).notifier).load();
      }
    });

    if (state.loading && state.detail == null) {
      return Scaffold(
        appBar: AppBar(),
        body: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Season skeleton
            const SkeletonLoader(height: 160, borderRadius: 16),
            const SizedBox(height: 20),
            // Members skeleton
            const SkeletonLoader(width: 120, height: 22, borderRadius: 6),
            const SizedBox(height: 12),
            ...List.generate(5, (_) => const MemberAvatarSkeleton()),
          ],
        ),
      );
    }

    if (state.error != null && state.detail == null) {
      return Scaffold(
        appBar: AppBar(),
        body: ErrorStateWidget(
          message: state.error,
          onRetry: () => ref
              .read(groupDetailProvider(widget.groupId).notifier)
              .load(),
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
          if (season != null && season.status == 'VOTING')
            Padding(
              padding: const EdgeInsets.only(right: 4),
              child: RevealCountdownWidget(revealAt: season.revealAt),
            ),
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
              const SizedBox(height: 12),
              if (QuestionVoteNotifier.isVotingWindowOpen())
                SizedBox(
                  width: double.infinity,
                  height: 44,
                  child: OutlinedButton.icon(
                    onPressed: () {
                      HapticFeedback.lightImpact();
                      context
                          .push('/groups/${widget.groupId}/question-vote');
                    },
                    icon: const Text('\u{1F5F3}',
                        style: TextStyle(fontSize: 18)),
                    label: const Text('Выбери вопрос недели'),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: AppColors.primary,
                      side: const BorderSide(color: AppColors.primary),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                  ),
                )
                    .animate(onPlay: (c) => c.repeat(reverse: true))
                    .scaleXY(
                      begin: 1.0,
                      end: 1.03,
                      duration: 1200.ms,
                      curve: Curves.easeInOut,
                    ),
              const SizedBox(height: 20),
            ],

            // Members
            Text('Участники', style: AppTextStyles.headline2),
            const SizedBox(height: 12),
            if (members.length <= 1)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 24),
                child: EmptyStateWidget(
                  emoji: '\u{1F517}',
                  title: 'Пока мало участников',
                  subtitle: 'Поделись ссылкой с друзьями',
                  buttonText: 'Пригласить',
                  onButtonPressed: _shareInvite,
                ),
              ),
            ...members.map(
              (m) => Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: InkWell(
                  borderRadius: BorderRadius.circular(8),
                  onTap: () {
                    HapticFeedback.lightImpact();
                    context.push('/groups/${widget.groupId}/members/${m.id}');
                  },
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 4),
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
            ),
          ],
        ),
      ),
    );
  }
}
