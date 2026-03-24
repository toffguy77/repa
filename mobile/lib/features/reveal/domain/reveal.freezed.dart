// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'reveal.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

RevealData _$RevealDataFromJson(Map<String, dynamic> json) {
  return _RevealData.fromJson(json);
}

/// @nodoc
mixin _$RevealData {
  @JsonKey(name: 'my_card')
  MyCard get myCard => throw _privateConstructorUsedError;
  @JsonKey(name: 'group_summary')
  GroupSummary get groupSummary => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $RevealDataCopyWith<RevealData> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $RevealDataCopyWith<$Res> {
  factory $RevealDataCopyWith(
          RevealData value, $Res Function(RevealData) then) =
      _$RevealDataCopyWithImpl<$Res, RevealData>;
  @useResult
  $Res call(
      {@JsonKey(name: 'my_card') MyCard myCard,
      @JsonKey(name: 'group_summary') GroupSummary groupSummary});

  $MyCardCopyWith<$Res> get myCard;
  $GroupSummaryCopyWith<$Res> get groupSummary;
}

/// @nodoc
class _$RevealDataCopyWithImpl<$Res, $Val extends RevealData>
    implements $RevealDataCopyWith<$Res> {
  _$RevealDataCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? myCard = null,
    Object? groupSummary = null,
  }) {
    return _then(_value.copyWith(
      myCard: null == myCard
          ? _value.myCard
          : myCard // ignore: cast_nullable_to_non_nullable
              as MyCard,
      groupSummary: null == groupSummary
          ? _value.groupSummary
          : groupSummary // ignore: cast_nullable_to_non_nullable
              as GroupSummary,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $MyCardCopyWith<$Res> get myCard {
    return $MyCardCopyWith<$Res>(_value.myCard, (value) {
      return _then(_value.copyWith(myCard: value) as $Val);
    });
  }

  @override
  @pragma('vm:prefer-inline')
  $GroupSummaryCopyWith<$Res> get groupSummary {
    return $GroupSummaryCopyWith<$Res>(_value.groupSummary, (value) {
      return _then(_value.copyWith(groupSummary: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$RevealDataImplCopyWith<$Res>
    implements $RevealDataCopyWith<$Res> {
  factory _$$RevealDataImplCopyWith(
          _$RevealDataImpl value, $Res Function(_$RevealDataImpl) then) =
      __$$RevealDataImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'my_card') MyCard myCard,
      @JsonKey(name: 'group_summary') GroupSummary groupSummary});

  @override
  $MyCardCopyWith<$Res> get myCard;
  @override
  $GroupSummaryCopyWith<$Res> get groupSummary;
}

/// @nodoc
class __$$RevealDataImplCopyWithImpl<$Res>
    extends _$RevealDataCopyWithImpl<$Res, _$RevealDataImpl>
    implements _$$RevealDataImplCopyWith<$Res> {
  __$$RevealDataImplCopyWithImpl(
      _$RevealDataImpl _value, $Res Function(_$RevealDataImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? myCard = null,
    Object? groupSummary = null,
  }) {
    return _then(_$RevealDataImpl(
      myCard: null == myCard
          ? _value.myCard
          : myCard // ignore: cast_nullable_to_non_nullable
              as MyCard,
      groupSummary: null == groupSummary
          ? _value.groupSummary
          : groupSummary // ignore: cast_nullable_to_non_nullable
              as GroupSummary,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$RevealDataImpl implements _RevealData {
  const _$RevealDataImpl(
      {@JsonKey(name: 'my_card') required this.myCard,
      @JsonKey(name: 'group_summary') required this.groupSummary});

  factory _$RevealDataImpl.fromJson(Map<String, dynamic> json) =>
      _$$RevealDataImplFromJson(json);

  @override
  @JsonKey(name: 'my_card')
  final MyCard myCard;
  @override
  @JsonKey(name: 'group_summary')
  final GroupSummary groupSummary;

  @override
  String toString() {
    return 'RevealData(myCard: $myCard, groupSummary: $groupSummary)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$RevealDataImpl &&
            (identical(other.myCard, myCard) || other.myCard == myCard) &&
            (identical(other.groupSummary, groupSummary) ||
                other.groupSummary == groupSummary));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, myCard, groupSummary);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$RevealDataImplCopyWith<_$RevealDataImpl> get copyWith =>
      __$$RevealDataImplCopyWithImpl<_$RevealDataImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$RevealDataImplToJson(
      this,
    );
  }
}

abstract class _RevealData implements RevealData {
  const factory _RevealData(
      {@JsonKey(name: 'my_card') required final MyCard myCard,
      @JsonKey(name: 'group_summary')
      required final GroupSummary groupSummary}) = _$RevealDataImpl;

  factory _RevealData.fromJson(Map<String, dynamic> json) =
      _$RevealDataImpl.fromJson;

  @override
  @JsonKey(name: 'my_card')
  MyCard get myCard;
  @override
  @JsonKey(name: 'group_summary')
  GroupSummary get groupSummary;
  @override
  @JsonKey(ignore: true)
  _$$RevealDataImplCopyWith<_$RevealDataImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

MyCard _$MyCardFromJson(Map<String, dynamic> json) {
  return _MyCard.fromJson(json);
}

/// @nodoc
mixin _$MyCard {
  @JsonKey(name: 'top_attributes')
  List<AttributeDto> get topAttributes => throw _privateConstructorUsedError;
  @JsonKey(name: 'hidden_attributes')
  List<AttributeDto> get hiddenAttributes => throw _privateConstructorUsedError;
  @JsonKey(name: 'reputation_title')
  String get reputationTitle => throw _privateConstructorUsedError;
  TrendDto? get trend => throw _privateConstructorUsedError;
  @JsonKey(name: 'new_achievements')
  List<AchievementDto> get newAchievements =>
      throw _privateConstructorUsedError;
  @JsonKey(name: 'card_image_url')
  String get cardImageUrl => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $MyCardCopyWith<MyCard> get copyWith => throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $MyCardCopyWith<$Res> {
  factory $MyCardCopyWith(MyCard value, $Res Function(MyCard) then) =
      _$MyCardCopyWithImpl<$Res, MyCard>;
  @useResult
  $Res call(
      {@JsonKey(name: 'top_attributes') List<AttributeDto> topAttributes,
      @JsonKey(name: 'hidden_attributes') List<AttributeDto> hiddenAttributes,
      @JsonKey(name: 'reputation_title') String reputationTitle,
      TrendDto? trend,
      @JsonKey(name: 'new_achievements') List<AchievementDto> newAchievements,
      @JsonKey(name: 'card_image_url') String cardImageUrl});

  $TrendDtoCopyWith<$Res>? get trend;
}

/// @nodoc
class _$MyCardCopyWithImpl<$Res, $Val extends MyCard>
    implements $MyCardCopyWith<$Res> {
  _$MyCardCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? topAttributes = null,
    Object? hiddenAttributes = null,
    Object? reputationTitle = null,
    Object? trend = freezed,
    Object? newAchievements = null,
    Object? cardImageUrl = null,
  }) {
    return _then(_value.copyWith(
      topAttributes: null == topAttributes
          ? _value.topAttributes
          : topAttributes // ignore: cast_nullable_to_non_nullable
              as List<AttributeDto>,
      hiddenAttributes: null == hiddenAttributes
          ? _value.hiddenAttributes
          : hiddenAttributes // ignore: cast_nullable_to_non_nullable
              as List<AttributeDto>,
      reputationTitle: null == reputationTitle
          ? _value.reputationTitle
          : reputationTitle // ignore: cast_nullable_to_non_nullable
              as String,
      trend: freezed == trend
          ? _value.trend
          : trend // ignore: cast_nullable_to_non_nullable
              as TrendDto?,
      newAchievements: null == newAchievements
          ? _value.newAchievements
          : newAchievements // ignore: cast_nullable_to_non_nullable
              as List<AchievementDto>,
      cardImageUrl: null == cardImageUrl
          ? _value.cardImageUrl
          : cardImageUrl // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $TrendDtoCopyWith<$Res>? get trend {
    if (_value.trend == null) {
      return null;
    }

    return $TrendDtoCopyWith<$Res>(_value.trend!, (value) {
      return _then(_value.copyWith(trend: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$MyCardImplCopyWith<$Res> implements $MyCardCopyWith<$Res> {
  factory _$$MyCardImplCopyWith(
          _$MyCardImpl value, $Res Function(_$MyCardImpl) then) =
      __$$MyCardImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'top_attributes') List<AttributeDto> topAttributes,
      @JsonKey(name: 'hidden_attributes') List<AttributeDto> hiddenAttributes,
      @JsonKey(name: 'reputation_title') String reputationTitle,
      TrendDto? trend,
      @JsonKey(name: 'new_achievements') List<AchievementDto> newAchievements,
      @JsonKey(name: 'card_image_url') String cardImageUrl});

  @override
  $TrendDtoCopyWith<$Res>? get trend;
}

/// @nodoc
class __$$MyCardImplCopyWithImpl<$Res>
    extends _$MyCardCopyWithImpl<$Res, _$MyCardImpl>
    implements _$$MyCardImplCopyWith<$Res> {
  __$$MyCardImplCopyWithImpl(
      _$MyCardImpl _value, $Res Function(_$MyCardImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? topAttributes = null,
    Object? hiddenAttributes = null,
    Object? reputationTitle = null,
    Object? trend = freezed,
    Object? newAchievements = null,
    Object? cardImageUrl = null,
  }) {
    return _then(_$MyCardImpl(
      topAttributes: null == topAttributes
          ? _value._topAttributes
          : topAttributes // ignore: cast_nullable_to_non_nullable
              as List<AttributeDto>,
      hiddenAttributes: null == hiddenAttributes
          ? _value._hiddenAttributes
          : hiddenAttributes // ignore: cast_nullable_to_non_nullable
              as List<AttributeDto>,
      reputationTitle: null == reputationTitle
          ? _value.reputationTitle
          : reputationTitle // ignore: cast_nullable_to_non_nullable
              as String,
      trend: freezed == trend
          ? _value.trend
          : trend // ignore: cast_nullable_to_non_nullable
              as TrendDto?,
      newAchievements: null == newAchievements
          ? _value._newAchievements
          : newAchievements // ignore: cast_nullable_to_non_nullable
              as List<AchievementDto>,
      cardImageUrl: null == cardImageUrl
          ? _value.cardImageUrl
          : cardImageUrl // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$MyCardImpl implements _MyCard {
  const _$MyCardImpl(
      {@JsonKey(name: 'top_attributes')
      required final List<AttributeDto> topAttributes,
      @JsonKey(name: 'hidden_attributes')
      required final List<AttributeDto> hiddenAttributes,
      @JsonKey(name: 'reputation_title') required this.reputationTitle,
      this.trend,
      @JsonKey(name: 'new_achievements')
      required final List<AchievementDto> newAchievements,
      @JsonKey(name: 'card_image_url') required this.cardImageUrl})
      : _topAttributes = topAttributes,
        _hiddenAttributes = hiddenAttributes,
        _newAchievements = newAchievements;

  factory _$MyCardImpl.fromJson(Map<String, dynamic> json) =>
      _$$MyCardImplFromJson(json);

  final List<AttributeDto> _topAttributes;
  @override
  @JsonKey(name: 'top_attributes')
  List<AttributeDto> get topAttributes {
    if (_topAttributes is EqualUnmodifiableListView) return _topAttributes;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_topAttributes);
  }

  final List<AttributeDto> _hiddenAttributes;
  @override
  @JsonKey(name: 'hidden_attributes')
  List<AttributeDto> get hiddenAttributes {
    if (_hiddenAttributes is EqualUnmodifiableListView)
      return _hiddenAttributes;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_hiddenAttributes);
  }

  @override
  @JsonKey(name: 'reputation_title')
  final String reputationTitle;
  @override
  final TrendDto? trend;
  final List<AchievementDto> _newAchievements;
  @override
  @JsonKey(name: 'new_achievements')
  List<AchievementDto> get newAchievements {
    if (_newAchievements is EqualUnmodifiableListView) return _newAchievements;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_newAchievements);
  }

  @override
  @JsonKey(name: 'card_image_url')
  final String cardImageUrl;

  @override
  String toString() {
    return 'MyCard(topAttributes: $topAttributes, hiddenAttributes: $hiddenAttributes, reputationTitle: $reputationTitle, trend: $trend, newAchievements: $newAchievements, cardImageUrl: $cardImageUrl)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$MyCardImpl &&
            const DeepCollectionEquality()
                .equals(other._topAttributes, _topAttributes) &&
            const DeepCollectionEquality()
                .equals(other._hiddenAttributes, _hiddenAttributes) &&
            (identical(other.reputationTitle, reputationTitle) ||
                other.reputationTitle == reputationTitle) &&
            (identical(other.trend, trend) || other.trend == trend) &&
            const DeepCollectionEquality()
                .equals(other._newAchievements, _newAchievements) &&
            (identical(other.cardImageUrl, cardImageUrl) ||
                other.cardImageUrl == cardImageUrl));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      const DeepCollectionEquality().hash(_topAttributes),
      const DeepCollectionEquality().hash(_hiddenAttributes),
      reputationTitle,
      trend,
      const DeepCollectionEquality().hash(_newAchievements),
      cardImageUrl);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$MyCardImplCopyWith<_$MyCardImpl> get copyWith =>
      __$$MyCardImplCopyWithImpl<_$MyCardImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$MyCardImplToJson(
      this,
    );
  }
}

abstract class _MyCard implements MyCard {
  const factory _MyCard(
      {@JsonKey(name: 'top_attributes')
      required final List<AttributeDto> topAttributes,
      @JsonKey(name: 'hidden_attributes')
      required final List<AttributeDto> hiddenAttributes,
      @JsonKey(name: 'reputation_title') required final String reputationTitle,
      final TrendDto? trend,
      @JsonKey(name: 'new_achievements')
      required final List<AchievementDto> newAchievements,
      @JsonKey(name: 'card_image_url')
      required final String cardImageUrl}) = _$MyCardImpl;

  factory _MyCard.fromJson(Map<String, dynamic> json) = _$MyCardImpl.fromJson;

  @override
  @JsonKey(name: 'top_attributes')
  List<AttributeDto> get topAttributes;
  @override
  @JsonKey(name: 'hidden_attributes')
  List<AttributeDto> get hiddenAttributes;
  @override
  @JsonKey(name: 'reputation_title')
  String get reputationTitle;
  @override
  TrendDto? get trend;
  @override
  @JsonKey(name: 'new_achievements')
  List<AchievementDto> get newAchievements;
  @override
  @JsonKey(name: 'card_image_url')
  String get cardImageUrl;
  @override
  @JsonKey(ignore: true)
  _$$MyCardImplCopyWith<_$MyCardImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

AttributeDto _$AttributeDtoFromJson(Map<String, dynamic> json) {
  return _AttributeDto.fromJson(json);
}

/// @nodoc
mixin _$AttributeDto {
  @JsonKey(name: 'question_id')
  String get questionId => throw _privateConstructorUsedError;
  @JsonKey(name: 'question_text')
  String get questionText => throw _privateConstructorUsedError;
  String get category => throw _privateConstructorUsedError;
  double get percentage => throw _privateConstructorUsedError;
  int get rank => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $AttributeDtoCopyWith<AttributeDto> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $AttributeDtoCopyWith<$Res> {
  factory $AttributeDtoCopyWith(
          AttributeDto value, $Res Function(AttributeDto) then) =
      _$AttributeDtoCopyWithImpl<$Res, AttributeDto>;
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'question_text') String questionText,
      String category,
      double percentage,
      int rank});
}

/// @nodoc
class _$AttributeDtoCopyWithImpl<$Res, $Val extends AttributeDto>
    implements $AttributeDtoCopyWith<$Res> {
  _$AttributeDtoCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? questionText = null,
    Object? category = null,
    Object? percentage = null,
    Object? rank = null,
  }) {
    return _then(_value.copyWith(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
      rank: null == rank
          ? _value.rank
          : rank // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$AttributeDtoImplCopyWith<$Res>
    implements $AttributeDtoCopyWith<$Res> {
  factory _$$AttributeDtoImplCopyWith(
          _$AttributeDtoImpl value, $Res Function(_$AttributeDtoImpl) then) =
      __$$AttributeDtoImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'question_text') String questionText,
      String category,
      double percentage,
      int rank});
}

/// @nodoc
class __$$AttributeDtoImplCopyWithImpl<$Res>
    extends _$AttributeDtoCopyWithImpl<$Res, _$AttributeDtoImpl>
    implements _$$AttributeDtoImplCopyWith<$Res> {
  __$$AttributeDtoImplCopyWithImpl(
      _$AttributeDtoImpl _value, $Res Function(_$AttributeDtoImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? questionText = null,
    Object? category = null,
    Object? percentage = null,
    Object? rank = null,
  }) {
    return _then(_$AttributeDtoImpl(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
      rank: null == rank
          ? _value.rank
          : rank // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$AttributeDtoImpl implements _AttributeDto {
  const _$AttributeDtoImpl(
      {@JsonKey(name: 'question_id') required this.questionId,
      @JsonKey(name: 'question_text') required this.questionText,
      required this.category,
      required this.percentage,
      required this.rank});

  factory _$AttributeDtoImpl.fromJson(Map<String, dynamic> json) =>
      _$$AttributeDtoImplFromJson(json);

  @override
  @JsonKey(name: 'question_id')
  final String questionId;
  @override
  @JsonKey(name: 'question_text')
  final String questionText;
  @override
  final String category;
  @override
  final double percentage;
  @override
  final int rank;

  @override
  String toString() {
    return 'AttributeDto(questionId: $questionId, questionText: $questionText, category: $category, percentage: $percentage, rank: $rank)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$AttributeDtoImpl &&
            (identical(other.questionId, questionId) ||
                other.questionId == questionId) &&
            (identical(other.questionText, questionText) ||
                other.questionText == questionText) &&
            (identical(other.category, category) ||
                other.category == category) &&
            (identical(other.percentage, percentage) ||
                other.percentage == percentage) &&
            (identical(other.rank, rank) || other.rank == rank));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, questionId, questionText, category, percentage, rank);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$AttributeDtoImplCopyWith<_$AttributeDtoImpl> get copyWith =>
      __$$AttributeDtoImplCopyWithImpl<_$AttributeDtoImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$AttributeDtoImplToJson(
      this,
    );
  }
}

abstract class _AttributeDto implements AttributeDto {
  const factory _AttributeDto(
      {@JsonKey(name: 'question_id') required final String questionId,
      @JsonKey(name: 'question_text') required final String questionText,
      required final String category,
      required final double percentage,
      required final int rank}) = _$AttributeDtoImpl;

  factory _AttributeDto.fromJson(Map<String, dynamic> json) =
      _$AttributeDtoImpl.fromJson;

  @override
  @JsonKey(name: 'question_id')
  String get questionId;
  @override
  @JsonKey(name: 'question_text')
  String get questionText;
  @override
  String get category;
  @override
  double get percentage;
  @override
  int get rank;
  @override
  @JsonKey(ignore: true)
  _$$AttributeDtoImplCopyWith<_$AttributeDtoImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

TrendDto _$TrendDtoFromJson(Map<String, dynamic> json) {
  return _TrendDto.fromJson(json);
}

/// @nodoc
mixin _$TrendDto {
  String get attribute => throw _privateConstructorUsedError;
  String get change => throw _privateConstructorUsedError;
  double get delta => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $TrendDtoCopyWith<TrendDto> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $TrendDtoCopyWith<$Res> {
  factory $TrendDtoCopyWith(TrendDto value, $Res Function(TrendDto) then) =
      _$TrendDtoCopyWithImpl<$Res, TrendDto>;
  @useResult
  $Res call({String attribute, String change, double delta});
}

/// @nodoc
class _$TrendDtoCopyWithImpl<$Res, $Val extends TrendDto>
    implements $TrendDtoCopyWith<$Res> {
  _$TrendDtoCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? attribute = null,
    Object? change = null,
    Object? delta = null,
  }) {
    return _then(_value.copyWith(
      attribute: null == attribute
          ? _value.attribute
          : attribute // ignore: cast_nullable_to_non_nullable
              as String,
      change: null == change
          ? _value.change
          : change // ignore: cast_nullable_to_non_nullable
              as String,
      delta: null == delta
          ? _value.delta
          : delta // ignore: cast_nullable_to_non_nullable
              as double,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$TrendDtoImplCopyWith<$Res>
    implements $TrendDtoCopyWith<$Res> {
  factory _$$TrendDtoImplCopyWith(
          _$TrendDtoImpl value, $Res Function(_$TrendDtoImpl) then) =
      __$$TrendDtoImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({String attribute, String change, double delta});
}

/// @nodoc
class __$$TrendDtoImplCopyWithImpl<$Res>
    extends _$TrendDtoCopyWithImpl<$Res, _$TrendDtoImpl>
    implements _$$TrendDtoImplCopyWith<$Res> {
  __$$TrendDtoImplCopyWithImpl(
      _$TrendDtoImpl _value, $Res Function(_$TrendDtoImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? attribute = null,
    Object? change = null,
    Object? delta = null,
  }) {
    return _then(_$TrendDtoImpl(
      attribute: null == attribute
          ? _value.attribute
          : attribute // ignore: cast_nullable_to_non_nullable
              as String,
      change: null == change
          ? _value.change
          : change // ignore: cast_nullable_to_non_nullable
              as String,
      delta: null == delta
          ? _value.delta
          : delta // ignore: cast_nullable_to_non_nullable
              as double,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$TrendDtoImpl implements _TrendDto {
  const _$TrendDtoImpl(
      {required this.attribute, required this.change, required this.delta});

  factory _$TrendDtoImpl.fromJson(Map<String, dynamic> json) =>
      _$$TrendDtoImplFromJson(json);

  @override
  final String attribute;
  @override
  final String change;
  @override
  final double delta;

  @override
  String toString() {
    return 'TrendDto(attribute: $attribute, change: $change, delta: $delta)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$TrendDtoImpl &&
            (identical(other.attribute, attribute) ||
                other.attribute == attribute) &&
            (identical(other.change, change) || other.change == change) &&
            (identical(other.delta, delta) || other.delta == delta));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, attribute, change, delta);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$TrendDtoImplCopyWith<_$TrendDtoImpl> get copyWith =>
      __$$TrendDtoImplCopyWithImpl<_$TrendDtoImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$TrendDtoImplToJson(
      this,
    );
  }
}

abstract class _TrendDto implements TrendDto {
  const factory _TrendDto(
      {required final String attribute,
      required final String change,
      required final double delta}) = _$TrendDtoImpl;

  factory _TrendDto.fromJson(Map<String, dynamic> json) =
      _$TrendDtoImpl.fromJson;

  @override
  String get attribute;
  @override
  String get change;
  @override
  double get delta;
  @override
  @JsonKey(ignore: true)
  _$$TrendDtoImplCopyWith<_$TrendDtoImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

AchievementDto _$AchievementDtoFromJson(Map<String, dynamic> json) {
  return _AchievementDto.fromJson(json);
}

/// @nodoc
mixin _$AchievementDto {
  String get type => throw _privateConstructorUsedError;
  Map<String, dynamic>? get metadata => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $AchievementDtoCopyWith<AchievementDto> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $AchievementDtoCopyWith<$Res> {
  factory $AchievementDtoCopyWith(
          AchievementDto value, $Res Function(AchievementDto) then) =
      _$AchievementDtoCopyWithImpl<$Res, AchievementDto>;
  @useResult
  $Res call({String type, Map<String, dynamic>? metadata});
}

/// @nodoc
class _$AchievementDtoCopyWithImpl<$Res, $Val extends AchievementDto>
    implements $AchievementDtoCopyWith<$Res> {
  _$AchievementDtoCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? type = null,
    Object? metadata = freezed,
  }) {
    return _then(_value.copyWith(
      type: null == type
          ? _value.type
          : type // ignore: cast_nullable_to_non_nullable
              as String,
      metadata: freezed == metadata
          ? _value.metadata
          : metadata // ignore: cast_nullable_to_non_nullable
              as Map<String, dynamic>?,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$AchievementDtoImplCopyWith<$Res>
    implements $AchievementDtoCopyWith<$Res> {
  factory _$$AchievementDtoImplCopyWith(_$AchievementDtoImpl value,
          $Res Function(_$AchievementDtoImpl) then) =
      __$$AchievementDtoImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({String type, Map<String, dynamic>? metadata});
}

/// @nodoc
class __$$AchievementDtoImplCopyWithImpl<$Res>
    extends _$AchievementDtoCopyWithImpl<$Res, _$AchievementDtoImpl>
    implements _$$AchievementDtoImplCopyWith<$Res> {
  __$$AchievementDtoImplCopyWithImpl(
      _$AchievementDtoImpl _value, $Res Function(_$AchievementDtoImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? type = null,
    Object? metadata = freezed,
  }) {
    return _then(_$AchievementDtoImpl(
      type: null == type
          ? _value.type
          : type // ignore: cast_nullable_to_non_nullable
              as String,
      metadata: freezed == metadata
          ? _value._metadata
          : metadata // ignore: cast_nullable_to_non_nullable
              as Map<String, dynamic>?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$AchievementDtoImpl implements _AchievementDto {
  const _$AchievementDtoImpl(
      {required this.type, final Map<String, dynamic>? metadata})
      : _metadata = metadata;

  factory _$AchievementDtoImpl.fromJson(Map<String, dynamic> json) =>
      _$$AchievementDtoImplFromJson(json);

  @override
  final String type;
  final Map<String, dynamic>? _metadata;
  @override
  Map<String, dynamic>? get metadata {
    final value = _metadata;
    if (value == null) return null;
    if (_metadata is EqualUnmodifiableMapView) return _metadata;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(value);
  }

  @override
  String toString() {
    return 'AchievementDto(type: $type, metadata: $metadata)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$AchievementDtoImpl &&
            (identical(other.type, type) || other.type == type) &&
            const DeepCollectionEquality().equals(other._metadata, _metadata));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, type, const DeepCollectionEquality().hash(_metadata));

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$AchievementDtoImplCopyWith<_$AchievementDtoImpl> get copyWith =>
      __$$AchievementDtoImplCopyWithImpl<_$AchievementDtoImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$AchievementDtoImplToJson(
      this,
    );
  }
}

abstract class _AchievementDto implements AchievementDto {
  const factory _AchievementDto(
      {required final String type,
      final Map<String, dynamic>? metadata}) = _$AchievementDtoImpl;

  factory _AchievementDto.fromJson(Map<String, dynamic> json) =
      _$AchievementDtoImpl.fromJson;

  @override
  String get type;
  @override
  Map<String, dynamic>? get metadata;
  @override
  @JsonKey(ignore: true)
  _$$AchievementDtoImplCopyWith<_$AchievementDtoImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

GroupSummary _$GroupSummaryFromJson(Map<String, dynamic> json) {
  return _GroupSummary.fromJson(json);
}

/// @nodoc
mixin _$GroupSummary {
  @JsonKey(name: 'top_per_question')
  List<TopQuestionResult> get topPerQuestion =>
      throw _privateConstructorUsedError;
  @JsonKey(name: 'voter_count')
  int get voterCount => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $GroupSummaryCopyWith<GroupSummary> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $GroupSummaryCopyWith<$Res> {
  factory $GroupSummaryCopyWith(
          GroupSummary value, $Res Function(GroupSummary) then) =
      _$GroupSummaryCopyWithImpl<$Res, GroupSummary>;
  @useResult
  $Res call(
      {@JsonKey(name: 'top_per_question')
      List<TopQuestionResult> topPerQuestion,
      @JsonKey(name: 'voter_count') int voterCount});
}

/// @nodoc
class _$GroupSummaryCopyWithImpl<$Res, $Val extends GroupSummary>
    implements $GroupSummaryCopyWith<$Res> {
  _$GroupSummaryCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? topPerQuestion = null,
    Object? voterCount = null,
  }) {
    return _then(_value.copyWith(
      topPerQuestion: null == topPerQuestion
          ? _value.topPerQuestion
          : topPerQuestion // ignore: cast_nullable_to_non_nullable
              as List<TopQuestionResult>,
      voterCount: null == voterCount
          ? _value.voterCount
          : voterCount // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$GroupSummaryImplCopyWith<$Res>
    implements $GroupSummaryCopyWith<$Res> {
  factory _$$GroupSummaryImplCopyWith(
          _$GroupSummaryImpl value, $Res Function(_$GroupSummaryImpl) then) =
      __$$GroupSummaryImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'top_per_question')
      List<TopQuestionResult> topPerQuestion,
      @JsonKey(name: 'voter_count') int voterCount});
}

/// @nodoc
class __$$GroupSummaryImplCopyWithImpl<$Res>
    extends _$GroupSummaryCopyWithImpl<$Res, _$GroupSummaryImpl>
    implements _$$GroupSummaryImplCopyWith<$Res> {
  __$$GroupSummaryImplCopyWithImpl(
      _$GroupSummaryImpl _value, $Res Function(_$GroupSummaryImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? topPerQuestion = null,
    Object? voterCount = null,
  }) {
    return _then(_$GroupSummaryImpl(
      topPerQuestion: null == topPerQuestion
          ? _value._topPerQuestion
          : topPerQuestion // ignore: cast_nullable_to_non_nullable
              as List<TopQuestionResult>,
      voterCount: null == voterCount
          ? _value.voterCount
          : voterCount // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$GroupSummaryImpl implements _GroupSummary {
  const _$GroupSummaryImpl(
      {@JsonKey(name: 'top_per_question')
      required final List<TopQuestionResult> topPerQuestion,
      @JsonKey(name: 'voter_count') required this.voterCount})
      : _topPerQuestion = topPerQuestion;

  factory _$GroupSummaryImpl.fromJson(Map<String, dynamic> json) =>
      _$$GroupSummaryImplFromJson(json);

  final List<TopQuestionResult> _topPerQuestion;
  @override
  @JsonKey(name: 'top_per_question')
  List<TopQuestionResult> get topPerQuestion {
    if (_topPerQuestion is EqualUnmodifiableListView) return _topPerQuestion;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_topPerQuestion);
  }

  @override
  @JsonKey(name: 'voter_count')
  final int voterCount;

  @override
  String toString() {
    return 'GroupSummary(topPerQuestion: $topPerQuestion, voterCount: $voterCount)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$GroupSummaryImpl &&
            const DeepCollectionEquality()
                .equals(other._topPerQuestion, _topPerQuestion) &&
            (identical(other.voterCount, voterCount) ||
                other.voterCount == voterCount));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType,
      const DeepCollectionEquality().hash(_topPerQuestion), voterCount);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$GroupSummaryImplCopyWith<_$GroupSummaryImpl> get copyWith =>
      __$$GroupSummaryImplCopyWithImpl<_$GroupSummaryImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$GroupSummaryImplToJson(
      this,
    );
  }
}

abstract class _GroupSummary implements GroupSummary {
  const factory _GroupSummary(
          {@JsonKey(name: 'top_per_question')
          required final List<TopQuestionResult> topPerQuestion,
          @JsonKey(name: 'voter_count') required final int voterCount}) =
      _$GroupSummaryImpl;

  factory _GroupSummary.fromJson(Map<String, dynamic> json) =
      _$GroupSummaryImpl.fromJson;

  @override
  @JsonKey(name: 'top_per_question')
  List<TopQuestionResult> get topPerQuestion;
  @override
  @JsonKey(name: 'voter_count')
  int get voterCount;
  @override
  @JsonKey(ignore: true)
  _$$GroupSummaryImplCopyWith<_$GroupSummaryImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

TopQuestionResult _$TopQuestionResultFromJson(Map<String, dynamic> json) {
  return _TopQuestionResult.fromJson(json);
}

/// @nodoc
mixin _$TopQuestionResult {
  @JsonKey(name: 'question_id')
  String get questionId => throw _privateConstructorUsedError;
  @JsonKey(name: 'question_text')
  String get questionText => throw _privateConstructorUsedError;
  @JsonKey(name: 'user_id')
  String get userId => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  double get percentage => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $TopQuestionResultCopyWith<TopQuestionResult> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $TopQuestionResultCopyWith<$Res> {
  factory $TopQuestionResultCopyWith(
          TopQuestionResult value, $Res Function(TopQuestionResult) then) =
      _$TopQuestionResultCopyWithImpl<$Res, TopQuestionResult>;
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'question_text') String questionText,
      @JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      double percentage});
}

/// @nodoc
class _$TopQuestionResultCopyWithImpl<$Res, $Val extends TopQuestionResult>
    implements $TopQuestionResultCopyWith<$Res> {
  _$TopQuestionResultCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? questionText = null,
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? percentage = null,
  }) {
    return _then(_value.copyWith(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$TopQuestionResultImplCopyWith<$Res>
    implements $TopQuestionResultCopyWith<$Res> {
  factory _$$TopQuestionResultImplCopyWith(_$TopQuestionResultImpl value,
          $Res Function(_$TopQuestionResultImpl) then) =
      __$$TopQuestionResultImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'question_text') String questionText,
      @JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      double percentage});
}

/// @nodoc
class __$$TopQuestionResultImplCopyWithImpl<$Res>
    extends _$TopQuestionResultCopyWithImpl<$Res, _$TopQuestionResultImpl>
    implements _$$TopQuestionResultImplCopyWith<$Res> {
  __$$TopQuestionResultImplCopyWithImpl(_$TopQuestionResultImpl _value,
      $Res Function(_$TopQuestionResultImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? questionText = null,
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? percentage = null,
  }) {
    return _then(_$TopQuestionResultImpl(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$TopQuestionResultImpl implements _TopQuestionResult {
  const _$TopQuestionResultImpl(
      {@JsonKey(name: 'question_id') required this.questionId,
      @JsonKey(name: 'question_text') required this.questionText,
      @JsonKey(name: 'user_id') required this.userId,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      required this.percentage});

  factory _$TopQuestionResultImpl.fromJson(Map<String, dynamic> json) =>
      _$$TopQuestionResultImplFromJson(json);

  @override
  @JsonKey(name: 'question_id')
  final String questionId;
  @override
  @JsonKey(name: 'question_text')
  final String questionText;
  @override
  @JsonKey(name: 'user_id')
  final String userId;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  final double percentage;

  @override
  String toString() {
    return 'TopQuestionResult(questionId: $questionId, questionText: $questionText, userId: $userId, username: $username, avatarEmoji: $avatarEmoji, percentage: $percentage)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$TopQuestionResultImpl &&
            (identical(other.questionId, questionId) ||
                other.questionId == questionId) &&
            (identical(other.questionText, questionText) ||
                other.questionText == questionText) &&
            (identical(other.userId, userId) || other.userId == userId) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.percentage, percentage) ||
                other.percentage == percentage));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, questionId, questionText, userId,
      username, avatarEmoji, percentage);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$TopQuestionResultImplCopyWith<_$TopQuestionResultImpl> get copyWith =>
      __$$TopQuestionResultImplCopyWithImpl<_$TopQuestionResultImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$TopQuestionResultImplToJson(
      this,
    );
  }
}

abstract class _TopQuestionResult implements TopQuestionResult {
  const factory _TopQuestionResult(
      {@JsonKey(name: 'question_id') required final String questionId,
      @JsonKey(name: 'question_text') required final String questionText,
      @JsonKey(name: 'user_id') required final String userId,
      required final String username,
      @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
      required final double percentage}) = _$TopQuestionResultImpl;

  factory _TopQuestionResult.fromJson(Map<String, dynamic> json) =
      _$TopQuestionResultImpl.fromJson;

  @override
  @JsonKey(name: 'question_id')
  String get questionId;
  @override
  @JsonKey(name: 'question_text')
  String get questionText;
  @override
  @JsonKey(name: 'user_id')
  String get userId;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  double get percentage;
  @override
  @JsonKey(ignore: true)
  _$$TopQuestionResultImplCopyWith<_$TopQuestionResultImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

MemberCard _$MemberCardFromJson(Map<String, dynamic> json) {
  return _MemberCard.fromJson(json);
}

/// @nodoc
mixin _$MemberCard {
  @JsonKey(name: 'user_id')
  String get userId => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl => throw _privateConstructorUsedError;
  @JsonKey(name: 'top_attributes')
  List<AttributeDto> get topAttributes => throw _privateConstructorUsedError;
  @JsonKey(name: 'reputation_title')
  String get reputationTitle => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $MemberCardCopyWith<MemberCard> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $MemberCardCopyWith<$Res> {
  factory $MemberCardCopyWith(
          MemberCard value, $Res Function(MemberCard) then) =
      _$MemberCardCopyWithImpl<$Res, MemberCard>;
  @useResult
  $Res call(
      {@JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl,
      @JsonKey(name: 'top_attributes') List<AttributeDto> topAttributes,
      @JsonKey(name: 'reputation_title') String reputationTitle});
}

/// @nodoc
class _$MemberCardCopyWithImpl<$Res, $Val extends MemberCard>
    implements $MemberCardCopyWith<$Res> {
  _$MemberCardCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
    Object? topAttributes = null,
    Object? reputationTitle = null,
  }) {
    return _then(_value.copyWith(
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      topAttributes: null == topAttributes
          ? _value.topAttributes
          : topAttributes // ignore: cast_nullable_to_non_nullable
              as List<AttributeDto>,
      reputationTitle: null == reputationTitle
          ? _value.reputationTitle
          : reputationTitle // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$MemberCardImplCopyWith<$Res>
    implements $MemberCardCopyWith<$Res> {
  factory _$$MemberCardImplCopyWith(
          _$MemberCardImpl value, $Res Function(_$MemberCardImpl) then) =
      __$$MemberCardImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl,
      @JsonKey(name: 'top_attributes') List<AttributeDto> topAttributes,
      @JsonKey(name: 'reputation_title') String reputationTitle});
}

/// @nodoc
class __$$MemberCardImplCopyWithImpl<$Res>
    extends _$MemberCardCopyWithImpl<$Res, _$MemberCardImpl>
    implements _$$MemberCardImplCopyWith<$Res> {
  __$$MemberCardImplCopyWithImpl(
      _$MemberCardImpl _value, $Res Function(_$MemberCardImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
    Object? topAttributes = null,
    Object? reputationTitle = null,
  }) {
    return _then(_$MemberCardImpl(
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      topAttributes: null == topAttributes
          ? _value._topAttributes
          : topAttributes // ignore: cast_nullable_to_non_nullable
              as List<AttributeDto>,
      reputationTitle: null == reputationTitle
          ? _value.reputationTitle
          : reputationTitle // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$MemberCardImpl implements _MemberCard {
  const _$MemberCardImpl(
      {@JsonKey(name: 'user_id') required this.userId,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      @JsonKey(name: 'avatar_url') this.avatarUrl,
      @JsonKey(name: 'top_attributes')
      required final List<AttributeDto> topAttributes,
      @JsonKey(name: 'reputation_title') required this.reputationTitle})
      : _topAttributes = topAttributes;

  factory _$MemberCardImpl.fromJson(Map<String, dynamic> json) =>
      _$$MemberCardImplFromJson(json);

  @override
  @JsonKey(name: 'user_id')
  final String userId;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  final String? avatarUrl;
  final List<AttributeDto> _topAttributes;
  @override
  @JsonKey(name: 'top_attributes')
  List<AttributeDto> get topAttributes {
    if (_topAttributes is EqualUnmodifiableListView) return _topAttributes;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_topAttributes);
  }

  @override
  @JsonKey(name: 'reputation_title')
  final String reputationTitle;

  @override
  String toString() {
    return 'MemberCard(userId: $userId, username: $username, avatarEmoji: $avatarEmoji, avatarUrl: $avatarUrl, topAttributes: $topAttributes, reputationTitle: $reputationTitle)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$MemberCardImpl &&
            (identical(other.userId, userId) || other.userId == userId) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.avatarUrl, avatarUrl) ||
                other.avatarUrl == avatarUrl) &&
            const DeepCollectionEquality()
                .equals(other._topAttributes, _topAttributes) &&
            (identical(other.reputationTitle, reputationTitle) ||
                other.reputationTitle == reputationTitle));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      userId,
      username,
      avatarEmoji,
      avatarUrl,
      const DeepCollectionEquality().hash(_topAttributes),
      reputationTitle);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$MemberCardImplCopyWith<_$MemberCardImpl> get copyWith =>
      __$$MemberCardImplCopyWithImpl<_$MemberCardImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$MemberCardImplToJson(
      this,
    );
  }
}

abstract class _MemberCard implements MemberCard {
  const factory _MemberCard(
      {@JsonKey(name: 'user_id') required final String userId,
      required final String username,
      @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
      @JsonKey(name: 'avatar_url') final String? avatarUrl,
      @JsonKey(name: 'top_attributes')
      required final List<AttributeDto> topAttributes,
      @JsonKey(name: 'reputation_title')
      required final String reputationTitle}) = _$MemberCardImpl;

  factory _MemberCard.fromJson(Map<String, dynamic> json) =
      _$MemberCardImpl.fromJson;

  @override
  @JsonKey(name: 'user_id')
  String get userId;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl;
  @override
  @JsonKey(name: 'top_attributes')
  List<AttributeDto> get topAttributes;
  @override
  @JsonKey(name: 'reputation_title')
  String get reputationTitle;
  @override
  @JsonKey(ignore: true)
  _$$MemberCardImplCopyWith<_$MemberCardImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

DetectorResult _$DetectorResultFromJson(Map<String, dynamic> json) {
  return _DetectorResult.fromJson(json);
}

/// @nodoc
mixin _$DetectorResult {
  bool get purchased => throw _privateConstructorUsedError;
  List<VoterProfile> get voters => throw _privateConstructorUsedError;
  @JsonKey(name: 'crystal_balance')
  int get crystalBalance => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $DetectorResultCopyWith<DetectorResult> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $DetectorResultCopyWith<$Res> {
  factory $DetectorResultCopyWith(
          DetectorResult value, $Res Function(DetectorResult) then) =
      _$DetectorResultCopyWithImpl<$Res, DetectorResult>;
  @useResult
  $Res call(
      {bool purchased,
      List<VoterProfile> voters,
      @JsonKey(name: 'crystal_balance') int crystalBalance});
}

/// @nodoc
class _$DetectorResultCopyWithImpl<$Res, $Val extends DetectorResult>
    implements $DetectorResultCopyWith<$Res> {
  _$DetectorResultCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? purchased = null,
    Object? voters = null,
    Object? crystalBalance = null,
  }) {
    return _then(_value.copyWith(
      purchased: null == purchased
          ? _value.purchased
          : purchased // ignore: cast_nullable_to_non_nullable
              as bool,
      voters: null == voters
          ? _value.voters
          : voters // ignore: cast_nullable_to_non_nullable
              as List<VoterProfile>,
      crystalBalance: null == crystalBalance
          ? _value.crystalBalance
          : crystalBalance // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$DetectorResultImplCopyWith<$Res>
    implements $DetectorResultCopyWith<$Res> {
  factory _$$DetectorResultImplCopyWith(_$DetectorResultImpl value,
          $Res Function(_$DetectorResultImpl) then) =
      __$$DetectorResultImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {bool purchased,
      List<VoterProfile> voters,
      @JsonKey(name: 'crystal_balance') int crystalBalance});
}

/// @nodoc
class __$$DetectorResultImplCopyWithImpl<$Res>
    extends _$DetectorResultCopyWithImpl<$Res, _$DetectorResultImpl>
    implements _$$DetectorResultImplCopyWith<$Res> {
  __$$DetectorResultImplCopyWithImpl(
      _$DetectorResultImpl _value, $Res Function(_$DetectorResultImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? purchased = null,
    Object? voters = null,
    Object? crystalBalance = null,
  }) {
    return _then(_$DetectorResultImpl(
      purchased: null == purchased
          ? _value.purchased
          : purchased // ignore: cast_nullable_to_non_nullable
              as bool,
      voters: null == voters
          ? _value._voters
          : voters // ignore: cast_nullable_to_non_nullable
              as List<VoterProfile>,
      crystalBalance: null == crystalBalance
          ? _value.crystalBalance
          : crystalBalance // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$DetectorResultImpl implements _DetectorResult {
  const _$DetectorResultImpl(
      {required this.purchased,
      required final List<VoterProfile> voters,
      @JsonKey(name: 'crystal_balance') required this.crystalBalance})
      : _voters = voters;

  factory _$DetectorResultImpl.fromJson(Map<String, dynamic> json) =>
      _$$DetectorResultImplFromJson(json);

  @override
  final bool purchased;
  final List<VoterProfile> _voters;
  @override
  List<VoterProfile> get voters {
    if (_voters is EqualUnmodifiableListView) return _voters;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_voters);
  }

  @override
  @JsonKey(name: 'crystal_balance')
  final int crystalBalance;

  @override
  String toString() {
    return 'DetectorResult(purchased: $purchased, voters: $voters, crystalBalance: $crystalBalance)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$DetectorResultImpl &&
            (identical(other.purchased, purchased) ||
                other.purchased == purchased) &&
            const DeepCollectionEquality().equals(other._voters, _voters) &&
            (identical(other.crystalBalance, crystalBalance) ||
                other.crystalBalance == crystalBalance));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, purchased,
      const DeepCollectionEquality().hash(_voters), crystalBalance);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$DetectorResultImplCopyWith<_$DetectorResultImpl> get copyWith =>
      __$$DetectorResultImplCopyWithImpl<_$DetectorResultImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$DetectorResultImplToJson(
      this,
    );
  }
}

abstract class _DetectorResult implements DetectorResult {
  const factory _DetectorResult(
      {required final bool purchased,
      required final List<VoterProfile> voters,
      @JsonKey(name: 'crystal_balance')
      required final int crystalBalance}) = _$DetectorResultImpl;

  factory _DetectorResult.fromJson(Map<String, dynamic> json) =
      _$DetectorResultImpl.fromJson;

  @override
  bool get purchased;
  @override
  List<VoterProfile> get voters;
  @override
  @JsonKey(name: 'crystal_balance')
  int get crystalBalance;
  @override
  @JsonKey(ignore: true)
  _$$DetectorResultImplCopyWith<_$DetectorResultImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VoterProfile _$VoterProfileFromJson(Map<String, dynamic> json) {
  return _VoterProfile.fromJson(json);
}

/// @nodoc
mixin _$VoterProfile {
  String get id => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VoterProfileCopyWith<VoterProfile> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VoterProfileCopyWith<$Res> {
  factory $VoterProfileCopyWith(
          VoterProfile value, $Res Function(VoterProfile) then) =
      _$VoterProfileCopyWithImpl<$Res, VoterProfile>;
  @useResult
  $Res call(
      {String id,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl});
}

/// @nodoc
class _$VoterProfileCopyWithImpl<$Res, $Val extends VoterProfile>
    implements $VoterProfileCopyWith<$Res> {
  _$VoterProfileCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$VoterProfileImplCopyWith<$Res>
    implements $VoterProfileCopyWith<$Res> {
  factory _$$VoterProfileImplCopyWith(
          _$VoterProfileImpl value, $Res Function(_$VoterProfileImpl) then) =
      __$$VoterProfileImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl});
}

/// @nodoc
class __$$VoterProfileImplCopyWithImpl<$Res>
    extends _$VoterProfileCopyWithImpl<$Res, _$VoterProfileImpl>
    implements _$$VoterProfileImplCopyWith<$Res> {
  __$$VoterProfileImplCopyWithImpl(
      _$VoterProfileImpl _value, $Res Function(_$VoterProfileImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
  }) {
    return _then(_$VoterProfileImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VoterProfileImpl implements _VoterProfile {
  const _$VoterProfileImpl(
      {required this.id,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      @JsonKey(name: 'avatar_url') this.avatarUrl});

  factory _$VoterProfileImpl.fromJson(Map<String, dynamic> json) =>
      _$$VoterProfileImplFromJson(json);

  @override
  final String id;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  final String? avatarUrl;

  @override
  String toString() {
    return 'VoterProfile(id: $id, username: $username, avatarEmoji: $avatarEmoji, avatarUrl: $avatarUrl)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VoterProfileImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.avatarUrl, avatarUrl) ||
                other.avatarUrl == avatarUrl));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, id, username, avatarEmoji, avatarUrl);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VoterProfileImplCopyWith<_$VoterProfileImpl> get copyWith =>
      __$$VoterProfileImplCopyWithImpl<_$VoterProfileImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VoterProfileImplToJson(
      this,
    );
  }
}

abstract class _VoterProfile implements VoterProfile {
  const factory _VoterProfile(
          {required final String id,
          required final String username,
          @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
          @JsonKey(name: 'avatar_url') final String? avatarUrl}) =
      _$VoterProfileImpl;

  factory _VoterProfile.fromJson(Map<String, dynamic> json) =
      _$VoterProfileImpl.fromJson;

  @override
  String get id;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl;
  @override
  @JsonKey(ignore: true)
  _$$VoterProfileImplCopyWith<_$VoterProfileImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

CardUrlResult _$CardUrlResultFromJson(Map<String, dynamic> json) {
  return _CardUrlResult.fromJson(json);
}

/// @nodoc
mixin _$CardUrlResult {
  @JsonKey(name: 'image_url')
  String? get imageUrl => throw _privateConstructorUsedError;
  String get status => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $CardUrlResultCopyWith<CardUrlResult> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $CardUrlResultCopyWith<$Res> {
  factory $CardUrlResultCopyWith(
          CardUrlResult value, $Res Function(CardUrlResult) then) =
      _$CardUrlResultCopyWithImpl<$Res, CardUrlResult>;
  @useResult
  $Res call({@JsonKey(name: 'image_url') String? imageUrl, String status});
}

/// @nodoc
class _$CardUrlResultCopyWithImpl<$Res, $Val extends CardUrlResult>
    implements $CardUrlResultCopyWith<$Res> {
  _$CardUrlResultCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? imageUrl = freezed,
    Object? status = null,
  }) {
    return _then(_value.copyWith(
      imageUrl: freezed == imageUrl
          ? _value.imageUrl
          : imageUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$CardUrlResultImplCopyWith<$Res>
    implements $CardUrlResultCopyWith<$Res> {
  factory _$$CardUrlResultImplCopyWith(
          _$CardUrlResultImpl value, $Res Function(_$CardUrlResultImpl) then) =
      __$$CardUrlResultImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({@JsonKey(name: 'image_url') String? imageUrl, String status});
}

/// @nodoc
class __$$CardUrlResultImplCopyWithImpl<$Res>
    extends _$CardUrlResultCopyWithImpl<$Res, _$CardUrlResultImpl>
    implements _$$CardUrlResultImplCopyWith<$Res> {
  __$$CardUrlResultImplCopyWithImpl(
      _$CardUrlResultImpl _value, $Res Function(_$CardUrlResultImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? imageUrl = freezed,
    Object? status = null,
  }) {
    return _then(_$CardUrlResultImpl(
      imageUrl: freezed == imageUrl
          ? _value.imageUrl
          : imageUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$CardUrlResultImpl implements _CardUrlResult {
  const _$CardUrlResultImpl(
      {@JsonKey(name: 'image_url') this.imageUrl, required this.status});

  factory _$CardUrlResultImpl.fromJson(Map<String, dynamic> json) =>
      _$$CardUrlResultImplFromJson(json);

  @override
  @JsonKey(name: 'image_url')
  final String? imageUrl;
  @override
  final String status;

  @override
  String toString() {
    return 'CardUrlResult(imageUrl: $imageUrl, status: $status)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$CardUrlResultImpl &&
            (identical(other.imageUrl, imageUrl) ||
                other.imageUrl == imageUrl) &&
            (identical(other.status, status) || other.status == status));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, imageUrl, status);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$CardUrlResultImplCopyWith<_$CardUrlResultImpl> get copyWith =>
      __$$CardUrlResultImplCopyWithImpl<_$CardUrlResultImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$CardUrlResultImplToJson(
      this,
    );
  }
}

abstract class _CardUrlResult implements CardUrlResult {
  const factory _CardUrlResult(
      {@JsonKey(name: 'image_url') final String? imageUrl,
      required final String status}) = _$CardUrlResultImpl;

  factory _CardUrlResult.fromJson(Map<String, dynamic> json) =
      _$CardUrlResultImpl.fromJson;

  @override
  @JsonKey(name: 'image_url')
  String? get imageUrl;
  @override
  String get status;
  @override
  @JsonKey(ignore: true)
  _$$CardUrlResultImplCopyWith<_$CardUrlResultImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

ReactionCounts _$ReactionCountsFromJson(Map<String, dynamic> json) {
  return _ReactionCounts.fromJson(json);
}

/// @nodoc
mixin _$ReactionCounts {
  Map<String, int> get counts => throw _privateConstructorUsedError;
  @JsonKey(name: 'my_emoji')
  String? get myEmoji => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ReactionCountsCopyWith<ReactionCounts> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ReactionCountsCopyWith<$Res> {
  factory $ReactionCountsCopyWith(
          ReactionCounts value, $Res Function(ReactionCounts) then) =
      _$ReactionCountsCopyWithImpl<$Res, ReactionCounts>;
  @useResult
  $Res call(
      {Map<String, int> counts, @JsonKey(name: 'my_emoji') String? myEmoji});
}

/// @nodoc
class _$ReactionCountsCopyWithImpl<$Res, $Val extends ReactionCounts>
    implements $ReactionCountsCopyWith<$Res> {
  _$ReactionCountsCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? counts = null,
    Object? myEmoji = freezed,
  }) {
    return _then(_value.copyWith(
      counts: null == counts
          ? _value.counts
          : counts // ignore: cast_nullable_to_non_nullable
              as Map<String, int>,
      myEmoji: freezed == myEmoji
          ? _value.myEmoji
          : myEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ReactionCountsImplCopyWith<$Res>
    implements $ReactionCountsCopyWith<$Res> {
  factory _$$ReactionCountsImplCopyWith(_$ReactionCountsImpl value,
          $Res Function(_$ReactionCountsImpl) then) =
      __$$ReactionCountsImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {Map<String, int> counts, @JsonKey(name: 'my_emoji') String? myEmoji});
}

/// @nodoc
class __$$ReactionCountsImplCopyWithImpl<$Res>
    extends _$ReactionCountsCopyWithImpl<$Res, _$ReactionCountsImpl>
    implements _$$ReactionCountsImplCopyWith<$Res> {
  __$$ReactionCountsImplCopyWithImpl(
      _$ReactionCountsImpl _value, $Res Function(_$ReactionCountsImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? counts = null,
    Object? myEmoji = freezed,
  }) {
    return _then(_$ReactionCountsImpl(
      counts: null == counts
          ? _value._counts
          : counts // ignore: cast_nullable_to_non_nullable
              as Map<String, int>,
      myEmoji: freezed == myEmoji
          ? _value.myEmoji
          : myEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ReactionCountsImpl implements _ReactionCounts {
  const _$ReactionCountsImpl(
      {required final Map<String, int> counts,
      @JsonKey(name: 'my_emoji') this.myEmoji})
      : _counts = counts;

  factory _$ReactionCountsImpl.fromJson(Map<String, dynamic> json) =>
      _$$ReactionCountsImplFromJson(json);

  final Map<String, int> _counts;
  @override
  Map<String, int> get counts {
    if (_counts is EqualUnmodifiableMapView) return _counts;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(_counts);
  }

  @override
  @JsonKey(name: 'my_emoji')
  final String? myEmoji;

  @override
  String toString() {
    return 'ReactionCounts(counts: $counts, myEmoji: $myEmoji)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ReactionCountsImpl &&
            const DeepCollectionEquality().equals(other._counts, _counts) &&
            (identical(other.myEmoji, myEmoji) || other.myEmoji == myEmoji));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, const DeepCollectionEquality().hash(_counts), myEmoji);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ReactionCountsImplCopyWith<_$ReactionCountsImpl> get copyWith =>
      __$$ReactionCountsImplCopyWithImpl<_$ReactionCountsImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ReactionCountsImplToJson(
      this,
    );
  }
}

abstract class _ReactionCounts implements ReactionCounts {
  const factory _ReactionCounts(
      {required final Map<String, int> counts,
      @JsonKey(name: 'my_emoji') final String? myEmoji}) = _$ReactionCountsImpl;

  factory _ReactionCounts.fromJson(Map<String, dynamic> json) =
      _$ReactionCountsImpl.fromJson;

  @override
  Map<String, int> get counts;
  @override
  @JsonKey(name: 'my_emoji')
  String? get myEmoji;
  @override
  @JsonKey(ignore: true)
  _$$ReactionCountsImplCopyWith<_$ReactionCountsImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
