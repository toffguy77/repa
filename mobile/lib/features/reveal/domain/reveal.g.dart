// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'reveal.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$RevealDataImpl _$$RevealDataImplFromJson(Map<String, dynamic> json) =>
    _$RevealDataImpl(
      myCard: MyCard.fromJson(json['my_card'] as Map<String, dynamic>),
      groupSummary:
          GroupSummary.fromJson(json['group_summary'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$RevealDataImplToJson(_$RevealDataImpl instance) =>
    <String, dynamic>{
      'my_card': instance.myCard,
      'group_summary': instance.groupSummary,
    };

_$MyCardImpl _$$MyCardImplFromJson(Map<String, dynamic> json) => _$MyCardImpl(
      topAttributes: (json['top_attributes'] as List<dynamic>)
          .map((e) => AttributeDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      hiddenAttributes: (json['hidden_attributes'] as List<dynamic>)
          .map((e) => AttributeDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      reputationTitle: json['reputation_title'] as String,
      trend: json['trend'] == null
          ? null
          : TrendDto.fromJson(json['trend'] as Map<String, dynamic>),
      newAchievements: (json['new_achievements'] as List<dynamic>)
          .map((e) => AchievementDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      cardImageUrl: json['card_image_url'] as String,
    );

Map<String, dynamic> _$$MyCardImplToJson(_$MyCardImpl instance) =>
    <String, dynamic>{
      'top_attributes': instance.topAttributes,
      'hidden_attributes': instance.hiddenAttributes,
      'reputation_title': instance.reputationTitle,
      'trend': instance.trend,
      'new_achievements': instance.newAchievements,
      'card_image_url': instance.cardImageUrl,
    };

_$AttributeDtoImpl _$$AttributeDtoImplFromJson(Map<String, dynamic> json) =>
    _$AttributeDtoImpl(
      questionId: json['question_id'] as String,
      questionText: json['question_text'] as String,
      category: json['category'] as String,
      percentage: (json['percentage'] as num).toDouble(),
      rank: (json['rank'] as num).toInt(),
    );

Map<String, dynamic> _$$AttributeDtoImplToJson(_$AttributeDtoImpl instance) =>
    <String, dynamic>{
      'question_id': instance.questionId,
      'question_text': instance.questionText,
      'category': instance.category,
      'percentage': instance.percentage,
      'rank': instance.rank,
    };

_$TrendDtoImpl _$$TrendDtoImplFromJson(Map<String, dynamic> json) =>
    _$TrendDtoImpl(
      attribute: json['attribute'] as String,
      change: json['change'] as String,
      delta: (json['delta'] as num).toDouble(),
    );

Map<String, dynamic> _$$TrendDtoImplToJson(_$TrendDtoImpl instance) =>
    <String, dynamic>{
      'attribute': instance.attribute,
      'change': instance.change,
      'delta': instance.delta,
    };

_$AchievementDtoImpl _$$AchievementDtoImplFromJson(Map<String, dynamic> json) =>
    _$AchievementDtoImpl(
      type: json['type'] as String,
      metadata: json['metadata'] as Map<String, dynamic>?,
    );

Map<String, dynamic> _$$AchievementDtoImplToJson(
        _$AchievementDtoImpl instance) =>
    <String, dynamic>{
      'type': instance.type,
      'metadata': instance.metadata,
    };

_$GroupSummaryImpl _$$GroupSummaryImplFromJson(Map<String, dynamic> json) =>
    _$GroupSummaryImpl(
      topPerQuestion: (json['top_per_question'] as List<dynamic>)
          .map((e) => TopQuestionResult.fromJson(e as Map<String, dynamic>))
          .toList(),
      voterCount: (json['voter_count'] as num).toInt(),
    );

Map<String, dynamic> _$$GroupSummaryImplToJson(_$GroupSummaryImpl instance) =>
    <String, dynamic>{
      'top_per_question': instance.topPerQuestion,
      'voter_count': instance.voterCount,
    };

_$TopQuestionResultImpl _$$TopQuestionResultImplFromJson(
        Map<String, dynamic> json) =>
    _$TopQuestionResultImpl(
      questionId: json['question_id'] as String,
      questionText: json['question_text'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      percentage: (json['percentage'] as num).toDouble(),
    );

Map<String, dynamic> _$$TopQuestionResultImplToJson(
        _$TopQuestionResultImpl instance) =>
    <String, dynamic>{
      'question_id': instance.questionId,
      'question_text': instance.questionText,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'percentage': instance.percentage,
    };

_$MemberCardImpl _$$MemberCardImplFromJson(Map<String, dynamic> json) =>
    _$MemberCardImpl(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      avatarUrl: json['avatar_url'] as String?,
      topAttributes: (json['top_attributes'] as List<dynamic>)
          .map((e) => AttributeDto.fromJson(e as Map<String, dynamic>))
          .toList(),
      reputationTitle: json['reputation_title'] as String,
    );

Map<String, dynamic> _$$MemberCardImplToJson(_$MemberCardImpl instance) =>
    <String, dynamic>{
      'user_id': instance.userId,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'avatar_url': instance.avatarUrl,
      'top_attributes': instance.topAttributes,
      'reputation_title': instance.reputationTitle,
    };

_$DetectorResultImpl _$$DetectorResultImplFromJson(Map<String, dynamic> json) =>
    _$DetectorResultImpl(
      purchased: json['purchased'] as bool,
      voters: (json['voters'] as List<dynamic>)
          .map((e) => VoterProfile.fromJson(e as Map<String, dynamic>))
          .toList(),
      crystalBalance: (json['crystal_balance'] as num).toInt(),
    );

Map<String, dynamic> _$$DetectorResultImplToJson(
        _$DetectorResultImpl instance) =>
    <String, dynamic>{
      'purchased': instance.purchased,
      'voters': instance.voters,
      'crystal_balance': instance.crystalBalance,
    };

_$VoterProfileImpl _$$VoterProfileImplFromJson(Map<String, dynamic> json) =>
    _$VoterProfileImpl(
      id: json['id'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      avatarUrl: json['avatar_url'] as String?,
    );

Map<String, dynamic> _$$VoterProfileImplToJson(_$VoterProfileImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'avatar_url': instance.avatarUrl,
    };

_$CardUrlResultImpl _$$CardUrlResultImplFromJson(Map<String, dynamic> json) =>
    _$CardUrlResultImpl(
      imageUrl: json['image_url'] as String?,
      status: json['status'] as String,
    );

Map<String, dynamic> _$$CardUrlResultImplToJson(_$CardUrlResultImpl instance) =>
    <String, dynamic>{
      'image_url': instance.imageUrl,
      'status': instance.status,
    };

_$ReactionCountsImpl _$$ReactionCountsImplFromJson(Map<String, dynamic> json) =>
    _$ReactionCountsImpl(
      counts: Map<String, int>.from(json['counts'] as Map),
      myEmoji: json['my_emoji'] as String?,
    );

Map<String, dynamic> _$$ReactionCountsImplToJson(
        _$ReactionCountsImpl instance) =>
    <String, dynamic>{
      'counts': instance.counts,
      'my_emoji': instance.myEmoji,
    };
