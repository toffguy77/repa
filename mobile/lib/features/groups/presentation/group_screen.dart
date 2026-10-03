import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/providers/auth_provider.dart';
import '../../../core/providers/connectivity_provider.dart';
import '../../../core/widgets/empty_state_widget.dart';
import '../../../core/widgets/error_state_widget.dart';
import '../../../core/widgets/reveal_countdown_widget.dart';
import '../../../core/widgets/skeleton_loader.dart';
import '../../question_vote/presentation/question_vote_notifier.dart';
import 'groups_notifier.dart';
import '../domain/group.dart';
import 'widgets/member_avatar.dart';
import 'widgets/pending_reveal_notice.dart';
import 'widgets/small_group_notice.dart';
import 'widgets/growth_notice.dart';
import 'widgets/kind_only_tile.dart';
import '../../../core/theme/app_tokens.dart';
import '../../../core/widgets/kit/kit.dart';
import 'widgets/invite_share_sheet.dart';
import 'widgets/member_actions_sheet.dart';

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
    showAppSheet<void>(
      context: context,
      builder: (_) => InviteShareSheet(
        inviteCode: detail.group.inviteCode,
        emoji: '\u{1F346}',
      ),
    );
  }

  /// Block, report, and (for the admin) remove. Reached by long-pressing a member, so the ordinary
  /// tap still opens their profile.
  void _showMemberActions(Member member, bool isAdminViewer) {
    showAppSheet<void>(
      context: context,
      builder: (_) => MemberActionsSheet(
        member: member,
        isAdminViewer: isAdminViewer,
        onBlock: () => _runSafetyAction(
          () => ref.read(groupsRepositoryProvider).blockMember(member.id),
          'Участник заблокирован',
        ),
        onReport: () => _runSafetyAction(
          () => ref
              .read(groupsRepositoryProvider)
              .reportMember(member.id, groupId: widget.groupId),
          'Жалоба отправлена',
        ),
        onRemove: isAdminViewer
            ? () => _runSafetyAction(
                  () => ref
                      .read(groupsRepositoryProvider)
                      .removeMember(widget.groupId, member.id),
                  'Участник удалён',
                )
            : null,
      ),
    );
  }

  /// Runs a safety action, refreshes the group, and reports the outcome either way — these are
  /// consequential actions and silence would be the wrong feedback.
  Future<void> _runSafetyAction(Future<void> Function() action, String success) async {
    try {
      await action();
      if (!mounted) return;
      await ref.read(groupDetailProvider(widget.groupId).notifier).load();
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(success)));
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(ErrorStateWidget.friendlyMessage(e.toString()))),
      );
    }
  }

  /// Leaving, with the two kinds distinguished. Ordinary leaving is reversible because people leave
  /// groups by accident; a permanent departure is the one an invite cannot undo.
  void _confirmLeave() {
    showAppSheet<void>(
      context: context,
      builder: (sheetContext) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text('Выйти из группы?', style: context.ts.heading2),
          SizedBox(height: AppTokens.space.sm),
          Text(
            'Голоса, которые ты уже отдал, останутся — они часть чужих результатов.',
            style: context.ts.caption,
            textAlign: TextAlign.center,
          ),
          SizedBox(height: AppTokens.space.xl),
          AppListTile(
            leading: Icon(Icons.logout, size: 20, color: context.t.color.textSecondary),
            title: 'Выйти',
            subtitle: 'Можно вернуться по ссылке',
            onTap: () {
              Navigator.pop(sheetContext);
              _leave(permanent: false);
            },
          ),
          AppListTile(
            leading: Icon(Icons.block, size: 20, color: context.t.color.danger),
            title: 'Выйти навсегда',
            subtitle: 'Вернуться по ссылке не получится',
            onTap: () {
              Navigator.pop(sheetContext);
              _leave(permanent: true);
            },
          ),
        ],
      ),
    );
  }

  Future<void> _leave({required bool permanent}) async {
    HapticFeedback.mediumImpact();
    final ok = await ref
        .read(groupDetailProvider(widget.groupId).notifier)
        .leave(permanent: permanent);
    if (!mounted) return;

    if (ok) {
      ref.read(groupsListProvider.notifier).refresh();
      context.go('/home');
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Не удалось выйти из группы')),
    );
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
            icon: const Icon(Icons.history),
            onPressed: () => context.push('/groups/${widget.groupId}/chronicle'),
            tooltip: 'История группы',
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
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: _confirmLeave,
            tooltip: 'Выйти из группы',
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
            SmallGroupNotice(
              memberCount: detail.effectiveMemberCount > 0
                  ? detail.effectiveMemberCount
                  : detail.members.length,
            ),
            GrowthNotice(threshold: detail.nextThreshold),
            if (detail.nextThreshold != null) const SizedBox(height: 12),
            KindOnlyTile(
              kindOnly: group.kindOnly,
              isAdmin: isAdmin,
              onChanged: (value) => ref
                  .read(groupDetailProvider(widget.groupId).notifier)
                  .setKindOnly(value),
            ),
            const SizedBox(height: 12),
            // Season status
            if (season != null) ...[
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
        color: context.t.elevation.level1.surface,
        borderRadius: AppTokens.radius.card,
        border: Border.all(
          color: context.t.elevation.level1.outline,
          width: AppTokens.border.hairline,
        ),
        boxShadow: context.t.elevation.level1.shadows,
      ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Сезон',
                      style: context.ts.heading2.copyWith(fontSize: 18),
                    ),
                    const SizedBox(height: 12),
                    if (season.status == 'VOTING') ...[
                      ClipRRect(
                        borderRadius: BorderRadius.circular(AppTokens.radius.xs),
                        child: LinearProgressIndicator(
                          value: season.totalCount > 0
                              ? season.votedCount / season.totalCount
                              : 0,
                          backgroundColor: context.t.color.surface,
                          color: context.t.color.accent,
                          minHeight: 8,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '${season.votedCount} из ${season.totalCount} проголосовали',
                        style: context.ts.caption,
                      ),
                      const SizedBox(height: 12),
                      PendingRevealNotice(season: season),
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
                                ? season.revealHeadline
                                : 'Проголосовать',
                          ),
                        ),
                      ),
                    ],
                    if (season.status == 'REVEALED') ...[
                      Text(
                        'Результаты готовы!',
                        style: context.ts.body.copyWith(
                          color: context.t.color.success,
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
                AppButton(
                  label: 'Выбери вопрос недели',
                  icon: Icons.how_to_vote_outlined,
                  variant: AppButtonVariant.secondary,
                  onPressed: () => context
                      .push('/groups/${widget.groupId}/question-vote'),
                )
                    .animate(onPlay: (c) => c.repeat(reverse: true))
                    .scaleXY(
                      begin: 1.0,
                      end: 1.03,
                      duration: context.motion(MotionClass.emphasis),
                      curve: Curves.easeInOut,
                    ),
              const SizedBox(height: 20),
            ],

            // Members
            Text('Участники', style: context.ts.heading2),
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
                  borderRadius: BorderRadius.circular(AppTokens.radius.sm),
                  onTap: () {
                    HapticFeedback.lightImpact();
                    context.push('/groups/${widget.groupId}/members/${m.id}');
                  },
                  onLongPress: m.id == currentUserId
                      ? null
                      : () {
                          HapticFeedback.mediumImpact();
                          _showMemberActions(m, isAdmin);
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
                          child: Text(m.username, style: context.ts.body),
                        ),
                        if (m.isAdmin)
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: context.t.color.accentFill.withValues(alpha: 0.18),
                              borderRadius: BorderRadius.circular(AppTokens.radius.sm),
                            ),
                            child: Text(
                              'Админ',
                              style: context.ts.caption.copyWith(
                                color: context.t.color.accent,
                                fontSize: 12,
                              ),
                            ),
                          ),
                        Icon(Icons.chevron_right, color: context.t.color.textSecondary),
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
