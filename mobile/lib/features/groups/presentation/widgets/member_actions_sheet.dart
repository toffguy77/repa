import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../domain/group.dart';

/// What a member can do about another member.
///
/// Exists because an anonymous peer-rating app for teenagers had no way to get away from anyone: the
/// only report was on a *question*, there was no block, and an admin could not remove anybody. App
/// Store Guideline 1.2 requires all three.
class MemberActionsSheet extends StatelessWidget {
  const MemberActionsSheet({
    super.key,
    required this.member,
    required this.isAdminViewer,
    required this.onBlock,
    required this.onReport,
    this.onRemove,
  });

  final Member member;

  /// Whether the *viewer* administers this group — removal is admin-only.
  final bool isAdminViewer;

  final VoidCallback onBlock;
  final VoidCallback onReport;
  final VoidCallback? onRemove;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(member.username, style: context.ts.heading2),
        SizedBox(height: AppTokens.space.xl),

        AppListTile(
          leading: Icon(Icons.block, size: 20, color: context.t.color.textSecondary),
          title: 'Заблокировать',
          subtitle: 'Вы исчезнете друг у друга из голосований и карточек',
          onTap: () => _confirm(
            context,
            title: 'Заблокировать ${member.username}?',
            body: 'Вы больше не сможете голосовать друг за друга и не увидите карточек друг друга. '
                'Это можно отменить.',
            confirmLabel: 'Заблокировать',
            onConfirm: onBlock,
          ),
        ),

        AppListTile(
          leading: Icon(Icons.flag_outlined, size: 20, color: context.t.color.warning),
          title: 'Пожаловаться',
          subtitle: 'Модераторы посмотрят. Участник не узнает',
          onTap: () => _confirm(
            context,
            title: 'Пожаловаться на ${member.username}?',
            body: 'Жалоба уйдёт модераторам. Участник не получит уведомления.',
            confirmLabel: 'Пожаловаться',
            onConfirm: onReport,
          ),
        ),

        if (isAdminViewer && onRemove != null)
          AppListTile(
            leading: Icon(Icons.person_remove_outlined, size: 20, color: context.t.color.danger),
            title: 'Удалить из группы',
            subtitle: 'Навсегда — по ссылке вернуться не получится',
            onTap: () => _confirm(
              context,
              title: 'Удалить ${member.username} из группы?',
              body: 'Участник потеряет доступ к группе и не сможет вернуться по инвайт-ссылке — '
                  'даже если создать новую. Отменить это нельзя.',
              confirmLabel: 'Удалить',
              danger: true,
              onConfirm: onRemove!,
            ),
          ),
      ],
    );
  }

  /// Every action states its consequence before it happens — these are not undoable by accident.
  void _confirm(
    BuildContext context, {
    required String title,
    required String body,
    required String confirmLabel,
    required VoidCallback onConfirm,
    bool danger = false,
  }) {
    showDialog<void>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(title),
        content: Text(body),
        actions: [
          AppButton(
            label: 'Отмена',
            variant: AppButtonVariant.ghost,
            expand: false,
            onPressed: () => Navigator.pop(dialogContext),
          ),
          AppButton(
            label: confirmLabel,
            danger: danger,
            expand: false,
            onPressed: () {
              HapticFeedback.mediumImpact();
              Navigator.pop(dialogContext);
              Navigator.maybePop(context);
              onConfirm();
            },
          ),
        ],
      ),
    );
  }
}
