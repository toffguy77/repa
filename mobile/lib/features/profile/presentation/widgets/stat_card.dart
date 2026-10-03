import 'package:flutter/material.dart';
import '../../../../core/theme/app_tokens.dart';

class StatCard extends StatefulWidget {
  final String label;
  final String value;
  final IconData icon;
  final bool animateNumber;

  const StatCard({
    super.key,
    required this.label,
    required this.value,
    required this.icon,
    this.animateNumber = true,
  });

  @override
  State<StatCard> createState() => _StatCardState();
}

class _StatCardState extends State<StatCard>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    // No duration yet: context.motion reads MediaQuery (for the reduce-motion preference), and
    // Flutter forbids reading an inherited widget before initState completes. Set in
    // didChangeDependencies below, which also means the duration follows the preference if it
    // changes while the card is on screen — the whole point of the motion token.
    _controller = AnimationController(vsync: this);
    _animation = CurvedAnimation(parent: _controller, curve: Curves.easeOut);

    if (widget.animateNumber) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) _controller.forward();
      });
    } else {
      _controller.value = 1.0;
    }
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _controller.duration = context.motion(MotionClass.emphasis);
  }

  @override
  void didUpdateWidget(StatCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.value != widget.value && widget.animateNumber) {
      _controller.forward(from: 0);
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final numValue = double.tryParse(
        widget.value.replaceAll('%', '').replaceAll(',', '.'));

    return Container(
      padding: const EdgeInsets.all(12),
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
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(widget.icon, size: 20, color: context.t.color.accent),
          const SizedBox(height: 8),
          if (numValue != null && widget.animateNumber)
            AnimatedBuilder(
              animation: _animation,
              builder: (context, _) {
                final current = numValue * _animation.value;
                final display = widget.value.contains('%')
                    ? '${current.toStringAsFixed(1)}%'
                    : current.toInt().toString();
                return Text(
                  display,
                  style: context.ts.heading2.copyWith(fontSize: 20),
                );
              },
            )
          else
            Text(
              widget.value,
              style: context.ts.heading2.copyWith(fontSize: 20),
            ),
          const SizedBox(height: 4),
          Text(
            widget.label,
            style: context.ts.caption.copyWith(fontSize: 12),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }
}

