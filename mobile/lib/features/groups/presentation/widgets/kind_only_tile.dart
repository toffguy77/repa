import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';

/// The group's kind-only setting, shown to everyone and changeable only by the admin.
///
/// Shown to non-admins too, read-only: the setting determines what members will be asked, so it is
/// information about the group rather than an admin preference. A switch that simply vanished for
/// most members would make the group's questions look arbitrary.
class KindOnlyTile extends StatefulWidget {
  const KindOnlyTile({
    super.key,
    required this.kindOnly,
    required this.isAdmin,
    required this.onChanged,
  });

  final bool kindOnly;
  final bool isAdmin;

  /// Returns the server's message when the change was refused, or null on success.
  final Future<String?> Function(bool value) onChanged;

  @override
  State<KindOnlyTile> createState() => _KindOnlyTileState();
}

class _KindOnlyTileState extends State<KindOnlyTile> {
  bool _saving = false;
  String? _error;

  Future<void> _change(bool value) async {
    HapticFeedback.selectionClick();
    setState(() {
      _saving = true;
      _error = null;
    });
    final error = await widget.onChanged(value);
    if (!mounted) return;
    setState(() {
      _saving = false;
      _error = error;
    });
  }

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text('Только добрые вопросы', style: context.ts.body),
              ),
              if (_saving)
                const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              else
                Switch(
                  value: widget.kindOnly,
                  // Null disables the switch, which is how a non-admin sees the setting without
                  // being able to change it.
                  onChanged: widget.isAdmin ? _change : null,
                ),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            widget.isAdmin
                // The timing is stated because it is the part that surprises people: the switch
                // looks like it should take effect now, and the open season deliberately does not
                // change under anyone who already answered.
                ? 'Колкие вопросы не будут приходить. Изменения применятся со следующей репы.'
                : 'Колкие вопросы в этой группе отключены. Менять может только админ.',
            style: context.ts.caption,
          ),
          if (_error != null) ...[
            const SizedBox(height: 8),
            Text(
              _error!,
              style: context.ts.caption.copyWith(color: context.t.color.danger),
            ),
          ],
        ],
      ),
    );
  }
}
